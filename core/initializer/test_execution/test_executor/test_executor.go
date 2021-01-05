/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor

import (
	"context"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/execution/exit_codes"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

/*
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

No logging to the system-level logger is allowed in this file!!! Everything should use the specific logger passed
	in at construction time, which allows us to capture per-test log messages so they don't all get jumbled together!

!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
 */

const (
	apiContainerLogFilename = "api-container.log"

	// The name of the directory inside a test execution directory where service file IO will be stored
	servicesDirname = "services"

	// When we're tearing down a network after a test (either after normal exit or test timeout), this is the maximum
	//  time we'll wait for each container to stop
	networkTeardownContainerStopTimeout = 10 * time.Second

	dockerSocket = "/var/run/docker.sock"

	testRunningContainerDescription = "Test-Running Container"

	networkNameTimestampFormat = "2006-01-02T15.04.05" // Go timestamp formatting is absolutely absurd...
)

/*
Because a test is run in its own goroutine to allow us to time it out, we need to pass the results back
	via a channel. This struct is what's passed over the channel.
 */
type testResult struct {
	// Whether the test passed or not (undefined if an error occurred that prevented us from retrieving test results)
	testPassed   bool

	// If not nil, the error that prevented us from retrieving the test result
	executionErr error
}


/*
Runs a single test with the given name

Args:
	executionInstanceId: The UUID representing an execution of the user's test suite, to which this test execution belongs
	ctx: The Context that the test execution is happening in
	log: the logger to which all logging events during test execution will be sent
	dockerClient: The Docker client to use to manipulate the Docker engine
	suiteExecutionVolume: The name of the Docker volume where file IO for the entire suite execution will be stored
	suiteExecutionVolumeDirpathOnInitializer: The dirpath, ON THE INITIALIZER CONTAINER, where the suite execution
		Docker volume is mounted
	testExecutionRelativeDirpath: The dirpath, relative to the root of the the suite execution volume, where file IO
		for this particular test should happen.

		NOTE: This directory must already exist!
	subnetMask: The subnet mask of the Docker network that has been spun up for this test
	kurtosisApiImageName: The name of the Docker image that will be used to run the Kurtosis API container
	apiContainerLogLevel: Log level that the Kurtosis API container should log at
	testsuiteLauncher: Launcher for running the test-running testsuite instances
	testsuiteDebuggerHostPortBinding: The port binding on the host machine that the testsuite debugger port should be tied to
	testName: The name of the test the executor should execute

Returns:
	bool: True if the test passed, false otherwise
	error: Non-nil if an error occurred that prevented the test pass/fail status from being retrieved
*/
func RunTest(
		executionInstanceId uuid.UUID,
		ctx context.Context,
		log *logrus.Logger,
		dockerClient *client.Client,
		suiteExecutionVolume string,
		suiteExecutionVolumeDirpathOnInitializer string,
		testExecutionRelativeDirpath string,
		subnetMask string,
		kurtosisApiImageName string,
		apiContainerLogLevel string,
		testsuiteLauncher *test_suite_constants.TestsuiteContainerLauncher,
		testsuiteDebuggerHostPortBinding nat.PortBinding,
		testName string) (bool, error) {
	log.Info("Creating Docker manager from environment settings...")
	// NOTE: at this point, all Docker commands from here forward will be bound by the Context that we pass in here - we'll
	//  only need to cancel this context once
	dockerManager, err := commons.NewDockerManager(log, dockerClient)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the Docker manager for test %v", testName)
	}
	log.Info("Docker manager created successfully")

	log.Infof("Creating Docker network for test with subnet mask %v...", subnetMask)
	freeIpAddrTracker, err := commons.NewFreeIpAddrTracker(
		log,
		subnetMask,
		map[string]bool{})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the free IP address tracker for test %v", testName)
	}
	gatewayIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting a free IP for the gateway for test %v", testName)
	}
	networkName := fmt.Sprintf(
		"%v_%v_%v",
		time.Now().Format(networkNameTimestampFormat),
		executionInstanceId.String(),
		testName)
	networkId, err := dockerManager.CreateNetwork(ctx, networkName, subnetMask, gatewayIp)
	if err != nil {
		// TODO If the user Ctrl-C's while the CreateNetwork call is ongoing then the CreateNetwork will error saying
		//  that the Context was cancelled as expected, but *the Docker engine will still create the networks!!! We'll
		//  need to parse the log message for the string "context canceled" and, if found, do another search for
		//  networks with our network name and delete them
		return false, stacktrace.Propagate(err, "Error occurred creating Docker network %v for test %v", networkName, testName)
	}
	defer removeNetworkDeferredFunc(log, dockerManager, networkId)
	log.Infof("Docker network %v created successfully", networkId)

	// TODO use hostnames rather than IPs, which makes things nicer and which we'll need for Docker swarm support
	// We need to create the IP addresses for BOTH containers because the testsuite needs to know the IP of the API
	//  container which will only be started after the testsuite container
	kurtosisApiIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}
	testRunningContainerIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting an IP for the test suite container running the test")
	}

	log.Debugf(
		"Test suite container IP: %v; kurtosis API container IP: %v",
		testRunningContainerIp.String(),
		kurtosisApiIp.String())

	servicesDirpathOnInitializerContainer := path.Join(
		suiteExecutionVolumeDirpathOnInitializer,
		testExecutionRelativeDirpath,
		servicesDirname)
	if err := os.Mkdir(servicesDirpathOnInitializerContainer, os.ModeDir); err != nil {
		return false, stacktrace.Propagate(
			err,
			"Could not create a directory inside the test execution directory, '%v', for storing services file IO",
			testExecutionRelativeDirpath)
	}

	servicesRelativeDirpath := path.Join(testExecutionRelativeDirpath, servicesDirname)

	log.Info("Launching testsuite container to run the test...")
	testRunningContainerId, err := testsuiteLauncher.LaunchTestRunningContainer(
		ctx,
		dockerManager,
		networkId,
		suiteExecutionVolume,
		testName,
		kurtosisApiIp.String(),
		testRunningContainerIp,
		servicesRelativeDirpath,
		testsuiteDebuggerHostPortBinding)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred launching the testsuite container to run the test")
	}
	log.Infof(
		"Test-running testsuite container launched, with debugger port bound to host port %v",
		testsuiteDebuggerHostPortBinding)
	log.Info("Successfully created test suite container to run the test")

	log.Info("Creating Kurtosis API container...")
	apiLogFilepathOnApiContainer := path.Join(
		api_container_docker_consts.SuiteExecutionVolumeMountDirpath,
		testExecutionRelativeDirpath,
		apiContainerLogFilename)
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/tcp", api_container_docker_consts.ContainerPort))
	kurtosisApiContainerId, err := dockerManager.CreateAndStartContainer(
		ctx,
		kurtosisApiImageName,
		networkId,
		kurtosisApiIp,
		map[commons.ContainerCapability]bool{},	// No extra capabilities needed for the API container
		commons.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			kurtosisApiPort: nil,
		},
		nil,
		map[string]string{
			api_container_env_vars.ApiContainerIpAddrEnvVar:       kurtosisApiIp.String(),
			// TODO capture the API container's logs ONLY if the user is an admin, so we don't leak internals
			//   about how our API container works to anyone trying to reverse-engineer Kurtosis
			api_container_env_vars.ApiLogFilepathEnvVar:           apiLogFilepathOnApiContainer,
			api_container_env_vars.GatewayIpEnvVar: gatewayIp.String(),
			// TODO Pull this value from testsuite metadata!!!
			api_container_env_vars.IsPartitioningEnabledEnvVar: "false",
			api_container_env_vars.LogLevelEnvVar: apiContainerLogLevel,
			api_container_env_vars.NetworkIdEnvVar: networkId,
			api_container_env_vars.SubnetMaskEnvVar: subnetMask,
			api_container_env_vars.TestSuiteContainerIdEnvVar: testRunningContainerId,
			api_container_env_vars.TestSuiteContainerIpAddrEnvVar: testRunningContainerIp.String(),
			api_container_env_vars.TestVolumeName: suiteExecutionVolume,
		},
		map[string]string{
			dockerSocket: dockerSocket,
		},
		map[string]string{
			suiteExecutionVolume: api_container_docker_consts.SuiteExecutionVolumeMountDirpath,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the Kurtosis API container")
	}
	log.Infof("Successfully created Kurtosis API container")

	// The Kurtosis API will be our indication of whether the test suite container stopped within the timeout or not
	log.Info("Waiting for Kurtosis API container to exit...")
	kurtosisApiExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		kurtosisApiContainerId)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred waiting for the exit of the Kurtosis API container: %v", err)
	}

	// At this point, we may be printing the logs of a stopped test suite container, or we may be printing the logs of
	//  still-running container that's exceeded the hard test timeout. Regardless, we want to print these so the user
	//  gets more information about what's going on, and the user will learn the exact error below
	banner_printer.PrintContainerLogsWithBanners(*dockerManager, ctx, testRunningContainerId, log, testRunningContainerDescription)

	var testStatusRetrievalError error
	switch kurtosisApiExitCode {
	case exit_codes.TestCompletedInTimeoutExitCode:
		testStatusRetrievalError = nil
	case exit_codes.OutOfOrderTestStatusExitCode:
		testStatusRetrievalError = stacktrace.NewError("The Kurtosis API container received an out-of-order " +
			"test execution status update; this is a Kurtosis code bug")
	case exit_codes.TestHitTimeoutExitCode:
		testStatusRetrievalError = stacktrace.NewError("The test failed to complete within the hard test " +
			"timeout (setup_buffer + test_execution_timeout), which most likely means the testnet setup took " +
			"too long (because if the test execution took too long, the test execution timeout" +
			"would have been tripped instead)")
	case exit_codes.NoTestSuiteRegisteredExitCode:
		testStatusRetrievalError = stacktrace.NewError("The test suite failed to register itself with the " +
			"Kurtosis API container; this is a bug with the test suite")
	case exit_codes.ShutdownSignalExitCode:
		testStatusRetrievalError = stacktrace.NewError("The Kurtosis API container exited due to receiving " +
			"a shutdown signal; if this is not expected, it's a Kurtosis bug")
	default:
		testStatusRetrievalError = stacktrace.NewError("The Kurtosis API container exited with an unrecognized " +
			"exit code %v; this is a code bug in Kurtosis indicating that the new exit code needs to be mapped " +
			"for the initializer", kurtosisApiExitCode)
	}
	if testStatusRetrievalError != nil {
		log.Error("An error occurred that prevented retrieval of the test completion status")
		return false, testStatusRetrievalError
	}
	log.Info("The test suite container running the test exited before the hard test timeout")

	// The test suite container will already have stopped, so now we get the exit code
	testSuiteExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		testRunningContainerId)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred retrieving the test suite container exit code")
	}
	return testSuiteExitCode == 0, nil
}


// =========================== PRIVATE HELPER FUNCTIONS =========================================

/*
Helper function for making a best-effort attempt at removing a network and logging any error states; intended to be run
as a deferred function.
*/
func removeNetworkDeferredFunc(log *logrus.Logger, dockerManager *commons.DockerManager, networkId string) {
	log.Infof("Attempting to remove Docker network with id %v...", networkId)
	// We use the background context here because we want to try and tear down the network even if the context the test was running in
	//  was cancelled. This might not be right - the right way to do it might be to pipe a separate context for the network teardown to here!
	if err := dockerManager.RemoveNetwork(context.Background(), networkId, networkTeardownContainerStopTimeout); err != nil {
		log.Errorf("An error occurred removing Docker network with ID %v:", networkId)
		log.Error(err.Error())
		log.Error("NOTE: This means you will need to clean up the Docker network manually!!")
	} else {
		log.Infof("Docker network with ID %v successfully removed", networkId)
	}
}
