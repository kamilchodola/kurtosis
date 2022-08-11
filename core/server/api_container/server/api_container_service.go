/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	kurtosis_backend_service "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/module_store"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	// Custom-set max size for logs coming back from docker exec.
	// Protobuf sets a maximum of 2GB for responses, in interest of keeping performance sane
	// we pick a reasonable limit of 10MB on log responses for docker exec.
	// See: https://stackoverflow.com/questions/34128872/google-protobuf-maximum-size/34186672
	maxLogOutputSizeBytes = 10 * 1024 * 1024

	// The string returned by the API if a service's public IP address doesn't exist
	missingPublicIpAddrStr = ""
	defaultContainerStopTimeoutSeconds = 0
)

// Guaranteed (by a unit test) to be a 1:1 mapping between API port protos and port spec protos
var apiContainerPortProtoToPortSpecPortProto = map[kurtosis_core_rpc_api_bindings.Port_Protocol]port_spec.PortProtocol{
	kurtosis_core_rpc_api_bindings.Port_TCP:  port_spec.PortProtocol_TCP,
	kurtosis_core_rpc_api_bindings.Port_SCTP: port_spec.PortProtocol_SCTP,
	kurtosis_core_rpc_api_bindings.Port_UDP:  port_spec.PortProtocol_UDP,
}

type ApiContainerService struct {
	// This embedding is required by gRPC
	kurtosis_core_rpc_api_bindings.UnimplementedApiContainerServiceServer

	filesArtifactStore *enclave_data_directory.FilesArtifactStore

	serviceNetwork *service_network.ServiceNetwork

	moduleStore *module_store.ModuleStore

	metricsClient client.MetricsClient
}

func NewApiContainerService(
	filesArtifactStore *enclave_data_directory.FilesArtifactStore,
	serviceNetwork *service_network.ServiceNetwork,
	moduleStore *module_store.ModuleStore,
	metricsClient client.MetricsClient,
) (*ApiContainerService, error) {
	service := &ApiContainerService{
		filesArtifactStore: filesArtifactStore,
		serviceNetwork:     serviceNetwork,
		moduleStore:        moduleStore,
		metricsClient:      metricsClient,
	}

	return service, nil
}

func (apicService ApiContainerService) LoadModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.LoadModuleArgs) (*kurtosis_core_rpc_api_bindings.LoadModuleResponse, error) {
	moduleId := module.ModuleID(args.ModuleId)
	image := args.ContainerImage
	serializedParams := args.SerializedParams

	if err := apicService.metricsClient.TrackLoadModule(args.ModuleId, image, serializedParams); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking load module event\n%v", err)
	}

	loadedModule, err := apicService.moduleStore.LoadModule(ctx, moduleId, image, serializedParams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred loading module '%v' with container image '%v' and serialized params '%v'", moduleId, image, serializedParams)
	}

	privateIpStr := loadedModule.GetPrivateIP().String()
	privateApiPort, err := transformPortSpecToApiPort(loadedModule.GetPrivatePort())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the module's private port spec port to an API port")
	}

	maybePublicIpStr := missingPublicIpAddrStr
	if loadedModule.GetMaybePublicIP() != nil {
		maybePublicIpStr = loadedModule.GetMaybePublicIP().String()
	}
	var maybePublicApiPort *kurtosis_core_rpc_api_bindings.Port
	if loadedModule.GetMaybePublicPort() != nil {
		candidatePublicApiPort, err := transformPortSpecToApiPort(loadedModule.GetMaybePublicPort())
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred transforming the module's public port '%v' to an API port",
				loadedModule.GetMaybePublicPort(),
			)
		}
		maybePublicApiPort = candidatePublicApiPort
	}

	result := binding_constructors.NewLoadModuleResponse(
		string(loadedModule.GetGUID()),
		privateIpStr,
		privateApiPort,
		maybePublicIpStr,
		maybePublicApiPort,
	)
	return result, nil
}

func (apicService ApiContainerService) UnloadModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UnloadModuleArgs) (*kurtosis_core_rpc_api_bindings.UnloadModuleResponse, error) {
	moduleId := module.ModuleID(args.ModuleId)

	if err := apicService.metricsClient.TrackUnloadModule(args.ModuleId); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking unload module event\n%v", err)
	}

	moduleGuid, err := apicService.moduleStore.UnloadModule(ctx, moduleId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unloading module '%v' from the network", moduleId)
	}

	return binding_constructors.NewUnloadModuleResponse(string(moduleGuid)), nil
}

func (apicService ApiContainerService) ExecuteModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecuteModuleArgs) (*kurtosis_core_rpc_api_bindings.ExecuteModuleResponse, error) {
	moduleId := module.ModuleID(args.ModuleId)
	serializedParams := args.SerializedParams

	if err := apicService.metricsClient.TrackExecuteModule(args.ModuleId, serializedParams); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking execute module event\n%v", err)
	}

	serializedResult, err := apicService.moduleStore.ExecuteModule(ctx, moduleId, serializedParams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing module '%v' with serialized params '%v'", moduleId, serializedParams)
	}

	resp := &kurtosis_core_rpc_api_bindings.ExecuteModuleResponse{SerializedResult: serializedResult}
	return resp, nil
}

func (apicService ApiContainerService) RegisterService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RegisterServiceArgs) (*kurtosis_core_rpc_api_bindings.RegisterServiceResponse, error) {
	serviceId := kurtosis_backend_service.ServiceID(args.ServiceId)
	partitionId := service_network_types.PartitionID(args.PartitionId)

	privateIpAddr, err := apicService.serviceNetwork.RegisterService(ctx, serviceId, partitionId)
	if err != nil {
		// TODO IP: Leaks internal information about API container
		return nil, stacktrace.Propagate(err, "An error occurred registering service '%v' in the service network", serviceId)
	}

	return binding_constructors.NewRegisterServiceResponse(privateIpAddr.String()), nil
}

func (apicService ApiContainerService) RegisterServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RegisterServicesArgs) (*kurtosis_core_rpc_api_bindings.RegisterServicesResponse, error) {
	serviceIDs := map[kurtosis_backend_service.ServiceID]bool{}
	for id := range args.ServiceIdSet {
		serviceIDs[kurtosis_backend_service.ServiceID(id)] = true
	}
	partitionId := service_network_types.PartitionID(args.PartitionId)
	serviceIDsToIPs, failedServiceErrors, err := apicService.serviceNetwork.RegisterServices(ctx, serviceIDs, partitionId)
	if err != nil {
		// TODO IP: Leaks internal information about API container
		return nil, stacktrace.Propagate(err, "An error occurred registering services '%v' in the service network", serviceIDs)
	}

	serviceIDsToIPsStrs := map[string]string{}
	for id, ip := range serviceIDsToIPs {
		serviceIDsToIPsStrs[string(id)] = ip.String()
	}
	failedServiceIDsToErrStrs := map[string]string{}
	for id, serviceErr := range failedServiceErrors {
		failedServiceIDsToErrStrs[string(id)] = serviceErr.Error()
	}

	return binding_constructors.NewRegisterServicesResponse(serviceIDsToIPsStrs, failedServiceIDsToErrStrs), nil
}

func (apicService ApiContainerService) StartService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StartServiceArgs) (*kurtosis_core_rpc_api_bindings.StartServiceResponse, error) {
	logrus.Debugf("Received request to start service with the following args: %+v", args)
	serviceId := kurtosis_backend_service.ServiceID(args.ServiceId)
	privateApiPorts := args.PrivatePorts

	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	requestedPublicApiPorts := args.PublicPorts
	if len(requestedPublicApiPorts) > 0 {

		if len(privateApiPorts) != len(requestedPublicApiPorts) {
			return nil, stacktrace.NewError("The received private ports length and the public ports length are not equal, received '%v' private ports and '%v' public ports", len(privateApiPorts), len(requestedPublicApiPorts))
		}

		for portId, privatePort := range privateApiPorts {
			if _, found := requestedPublicApiPorts[portId]; !found {
				return nil, stacktrace.NewError("Expected to receive public port with ID '%v' bound to private port number '%v', but it was not found", portId, privatePort.GetNumber())
			}
		}
	}

	requestedPublicServicePortSpecs := map[string]*port_spec.PortSpec{}
	for portId, publicApiPort := range requestedPublicApiPorts {
		publicServicePortSpec, err := transformApiPortToPortSpec(publicApiPort)
		if err != nil {
			return nil, stacktrace.NewError("An error occurred transforming the API port for public port '%v' into a port spec port", portId)
		}
		requestedPublicServicePortSpecs[portId] = publicServicePortSpec
	}
	//TODO Finished the huge hack to temporarily enable static ports for NEAR

	privateServicePortSpecs := map[string]*port_spec.PortSpec{}
	for portId, privateApiPort := range privateApiPorts {
		privateServicePortSpec, err := transformApiPortToPortSpec(privateApiPort)
		if err != nil {
			return nil, stacktrace.NewError("An error occurred transforming the API port for private port '%v' into a port spec port", portId)
		}
		privateServicePortSpecs[portId] = privateServicePortSpec
	}
	filesArtifactMountpointsByArtifactUuid := map[enclave_data_directory.FilesArtifactUUID]string{}
	for filesArtifactUuidStr, mountDirPath := range args.FilesArtifactMountpoints {
		filesArtifactMountpointsByArtifactUuid[enclave_data_directory.FilesArtifactUUID(filesArtifactUuidStr)] = mountDirPath
	}
	startedService, err := apicService.serviceNetwork.StartService(
		ctx,
		serviceId,
		args.DockerImage,
		privateServicePortSpecs,
		requestedPublicServicePortSpecs,
		args.EntrypointArgs,
		args.CmdArgs,
		args.DockerEnvVars,
		filesArtifactMountpointsByArtifactUuid,
		args.CpuAllocationMillicpus,
		args.MemoryAllocationMegabytes,
	)
	if err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred starting the service in the service network")
	}
	privateServiceIpStr := startedService.GetRegistration().GetPrivateIP().String()
	serviceGuidStr := string(startedService.GetRegistration().GetGUID())
	publicServicePortSpecs := startedService.GetMaybePublicPorts()
	publicApiPorts, err := transformPortSpecMapToApiPortsMap(publicServicePortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the service's public port specs to API ports")
	}
	maybePublicIpAddr := startedService.GetMaybePublicIP()
	publicIpAddrStr := missingPublicIpAddrStr
	if maybePublicIpAddr != nil {
		publicIpAddrStr = maybePublicIpAddr.String()
	}
	response := binding_constructors.NewStartServiceResponse(privateServiceIpStr, privateApiPorts, publicIpAddrStr, publicApiPorts, serviceGuidStr)

	serviceStartLoglineSuffix := ""
	if len(publicServicePortSpecs) > 0 {
		serviceStartLoglineSuffix = fmt.Sprintf(
			" with the following public ports: %+v",
			publicServicePortSpecs,
		)
	}
	logrus.Infof("Started service '%v'%v", serviceId, serviceStartLoglineSuffix)

	return response, nil
}

func (apicService ApiContainerService) StartServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StartServicesArgs) (*kurtosis_core_rpc_api_bindings.StartServicesResponse, error){
	failedServicesPool := map[kurtosis_backend_service.ServiceID]error{}
	serviceIDsToConfigs := map[kurtosis_backend_service.ServiceID]*kurtosis_backend_service.ServiceConfig{}
	serviceIDsToFilesArtifactUUIDsToMountpoints := map[kurtosis_backend_service.ServiceID]map[enclave_data_directory.FilesArtifactUUID]string{}
	for serviceIDStr, serviceConfig := range args.ServiceIdsToConfigs {
		serviceID := kurtosis_backend_service.ServiceID(serviceIDStr)
		logrus.Debugf("Received request to start service with the following args: %+v", serviceConfig)
		privateServicePortSpecs, requestedPublicServicePortSpecs, err := convertAPIPortsToPortSpecs(serviceConfig.PrivatePorts, serviceConfig.PublicPorts)
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred while trying to convert public and private API ports to port specs for service '%v'", serviceID)
			continue
		}

		filesArtifactMountpointsByArtifactUUID := map[enclave_data_directory.FilesArtifactUUID]string{}
		for filesArtifactUUIDStr, mountDirPath := range serviceConfig.FilesArtifactMountpoints {
			filesArtifactMountpointsByArtifactUUID[enclave_data_directory.FilesArtifactUUID(filesArtifactUUIDStr)] = mountDirPath
		}
		serviceIDsToFilesArtifactUUIDsToMountpoints[serviceID] = filesArtifactMountpointsByArtifactUUID

		serviceConfigObj := kurtosis_backend_service.NewServiceConfig(
			serviceConfig.ContainerImageName,
			privateServicePortSpecs,
			requestedPublicServicePortSpecs,
			serviceConfig.EntrypointArgs,
			serviceConfig.CmdArgs,
			serviceConfig.EnvVars,
			nil, // Will get set later if needed
			serviceConfig.CpuAllocationMillicpus,
			serviceConfig.MemoryAllocationMegabytes,
		)
		serviceIDsToConfigs[serviceID] = serviceConfigObj
	}

	successfulServices, failedServices, err := apicService.serviceNetwork.StartServices(ctx, serviceIDsToConfigs, serviceIDsToFilesArtifactUUIDsToMountpoints)
	if err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred starting services in the service network")
	}

	for serviceID, serviceErr := range failedServices {
		failedServicesPool[serviceID] = serviceErr
		logrus.Debugf("Failed to start service '%v'", serviceID)
	}

	serviceIDsToServiceInfo := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	for serviceID, startedService := range successfulServices {
		// If anything goes wrong while trying to set up the service object, we need to remove the successfully started service
		shouldRemoveService := true
		defer func() {
			if shouldRemoveService {
				_, err := apicService.serviceNetwork.RemoveService(ctx, serviceID, defaultContainerStopTimeoutSeconds)
				if err != nil {
					failedServicesPool[serviceID] = stacktrace.NewError(
						"WARNING: Attempted to remove service '%v' to delete its resources after it failed to create service object, but an error occurred" +
					"while attempting to remove the service. This means there exists a service that should not have been started!.", serviceID)
				}
			}
		}()

		serviceRegistration := startedService.GetRegistration()
		serviceGuidStr := string(serviceRegistration.GetGUID())
		privateServiceIpStr := serviceRegistration.GetPrivateIP().String()
		privateServicePortSpecs := startedService.GetPrivatePorts()
		privateApiPorts, err := transformPortSpecMapToApiPortsMap(privateServicePortSpecs)
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred transforming the service '%v' private port specs to API ports", serviceID)
			continue
		}
		publicServicePortSpecs := startedService.GetMaybePublicPorts()
		publicApiPorts, err := transformPortSpecMapToApiPortsMap(publicServicePortSpecs)
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred transforming the service '%v' public port specs to API ports.", serviceID)
			continue
		}
		maybePublicIpAddr := startedService.GetMaybePublicIP()
		publicIpAddrStr := missingPublicIpAddrStr
		if maybePublicIpAddr != nil {
			publicIpAddrStr = maybePublicIpAddr.String()
		}

		shouldRemoveService = false
		serviceIDsToServiceInfo[string(serviceID)] = binding_constructors.NewServiceInfo(serviceGuidStr, privateServiceIpStr, privateApiPorts, publicIpAddrStr, publicApiPorts)
		serviceStartLoglineSuffix := ""
		if len(publicServicePortSpecs) > 0 {
			serviceStartLoglineSuffix = fmt.Sprintf(
				" with the following public ports: %+v",
				publicServicePortSpecs,
			)
		}
		logrus.Infof("Started service '%v'%v", serviceID, serviceStartLoglineSuffix)
	}

	failedServiceIDsToErrorStr := map[string]string{}
	for id, serviceErr := range failedServicesPool {
		failedServiceIDsToErrorStr[string(id)] = serviceErr.Error()
	}
	return binding_constructors.NewStartServicesResponse(serviceIDsToServiceInfo, failedServiceIDsToErrorStr), nil
}

func (apicService ApiContainerService) RemoveService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RemoveServiceArgs) (*kurtosis_core_rpc_api_bindings.RemoveServiceResponse, error) {
	serviceId := kurtosis_backend_service.ServiceID(args.ServiceId)

	containerStopTimeoutSeconds := args.ContainerStopTimeoutSeconds
	containerStopTimeout := time.Duration(containerStopTimeoutSeconds) * time.Second

	serviceGuid, err := apicService.serviceNetwork.RemoveService(ctx, serviceId, containerStopTimeout)
	if err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred removing service with ID '%v'", serviceId)
	}
	return binding_constructors.NewRemoveServiceResponse(string(serviceGuid)), nil
}

func (apicService ApiContainerService) Repartition(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RepartitionArgs) (*emptypb.Empty, error) {
	// No need to check for dupes here - that happens at the lowest-level call to ServiceNetwork.Repartition (as it should)
	partitionServices := map[service_network_types.PartitionID]map[kurtosis_backend_service.ServiceID]bool{}
	for partitionIdStr, servicesInPartition := range args.PartitionServices {
		partitionId := service_network_types.PartitionID(partitionIdStr)
		serviceIdSet := map[kurtosis_backend_service.ServiceID]bool{}
		for serviceIdStr := range servicesInPartition.ServiceIdSet {
			serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
			serviceIdSet[serviceId] = true
		}
		partitionServices[partitionId] = serviceIdSet
	}

	partitionConnections := map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection{}
	for partitionAStr, partitionBToConnection := range args.PartitionConnections {
		partitionAId := service_network_types.PartitionID(partitionAStr)
		for partitionBStr, connectionInfo := range partitionBToConnection.ConnectionInfo {
			partitionBId := service_network_types.PartitionID(partitionBStr)
			partitionConnectionId := *service_network_types.NewPartitionConnectionID(partitionAId, partitionBId)
			if _, found := partitionConnections[partitionConnectionId]; found {
				return nil, stacktrace.NewError(
					"Partition connection '%v' <-> '%v' was defined twice (possibly in reverse order)",
					partitionAId,
					partitionBId)
			}
			partitionConnection := partition_topology.PartitionConnection{
				PacketLossPercentage: connectionInfo.PacketLossPercentage,
			}
			partitionConnections[partitionConnectionId] = partitionConnection
		}
	}

	defaultConnectionInfo := args.DefaultConnection
	defaultConnection := partition_topology.PartitionConnection{
		PacketLossPercentage: defaultConnectionInfo.PacketLossPercentage,
	}

	if err := apicService.serviceNetwork.Repartition(
		ctx,
		partitionServices,
		partitionConnections,
		defaultConnection); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred repartitioning the test network")
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) PauseService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.PauseServiceArgs) (*emptypb.Empty, error) {
	serviceIdStr := args.ServiceId
	serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
	err := service.serviceNetwork.PauseService(ctx, serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to pause service '%v'", serviceId)
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) UnpauseService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UnpauseServiceArgs) (*emptypb.Empty, error) {
	serviceIdStr := args.ServiceId
	serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
	err := service.serviceNetwork.UnpauseService(ctx, serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to unpause service '%v'", serviceId)
	}
	return &emptypb.Empty{}, nil
}

func (apicService ApiContainerService) ExecCommand(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecCommandArgs) (*kurtosis_core_rpc_api_bindings.ExecCommandResponse, error) {
	serviceIdStr := args.ServiceId
	serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
	command := args.CommandArgs
	exitCode, logOutput, err := apicService.serviceNetwork.ExecCommand(ctx, serviceId, command)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running exec command '%v' against service '%v' in the service network",
			command,
			serviceId)
	}
	numLogOutputBytes := len(logOutput)
	if numLogOutputBytes > maxLogOutputSizeBytes {
		return nil, stacktrace.NewError(
			"Log output from docker exec command '%+v' was %v bytes, but maximum size allowed by Kurtosis is %v",
			command,
			numLogOutputBytes,
			maxLogOutputSizeBytes,
		)
	}
	resp := &kurtosis_core_rpc_api_bindings.ExecCommandResponse{
		ExitCode:  exitCode,
		LogOutput: logOutput,
	}
	return resp, nil
}

func (apicService ApiContainerService) WaitForHttpGetEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs) (*emptypb.Empty, error) {

	serviceIdStr := args.ServiceId

	if err := apicService.waitForEndpointAvailability(
		ctx,
		serviceIdStr,
		http.MethodGet,
		args.Port,
		args.Path,
		args.InitialDelayMilliseconds,
		args.Retries,
		args.RetriesDelayMilliseconds,
		"",
		args.BodyText); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for HTTP endpoint '%v' to become available",
			args.Path,
		)
	}

	return &emptypb.Empty{}, nil
}

func (apicService ApiContainerService) WaitForHttpPostEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs) (*emptypb.Empty, error) {

	serviceIdStr := args.ServiceId

	if err := apicService.waitForEndpointAvailability(
		ctx,
		serviceIdStr,
		http.MethodPost,
		args.Port,
		args.Path,
		args.InitialDelayMilliseconds,
		args.Retries,
		args.RetriesDelayMilliseconds,
		args.RequestBody,
		args.BodyText); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for HTTP endpoint '%v' to become available",
			args.Path,
		)
	}

	return &emptypb.Empty{}, nil
}

func (apicService ApiContainerService) GetServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetServicesArgs) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error) {
	serviceInfos := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	filterServiceIds := args.ServiceIds

	for serviceID, _ := range apicService.serviceNetwork.GetServiceIDs() {
		serviceIDStr := string(serviceID)
		if filterServiceIds != nil && len(filterServiceIds) > 0 {
			if _, found := filterServiceIds[serviceIDStr]; !found {
				continue
			}
		}
		serviceInfo, err := apicService.getServiceInfo(ctx, serviceID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get service info for service '%v'", serviceID)
		}
		serviceInfos[serviceIDStr] = serviceInfo
	}

	resp := binding_constructors.NewGetServicesResponse(serviceInfos)
	return resp, nil
}

func (apicService ApiContainerService) GetModules(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetModulesArgs) (*kurtosis_core_rpc_api_bindings.GetModulesResponse, error) {
	moduleInfos := map[string]*kurtosis_core_rpc_api_bindings.ModuleInfo{}
	filterModuleIds := args.Ids

	for moduleID, _ := range apicService.moduleStore.GetModules() {
		moduleIDStr := string(moduleID)
		if filterModuleIds != nil && len(filterModuleIds) > 0 {
			if _, found := filterModuleIds[moduleIDStr]; !found {
				continue
			}
		}
		moduleInfo, err := apicService.getModuleInfo(ctx, moduleID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get Module info for Module '%v'", moduleID)
		}
		moduleInfos[moduleIDStr] = moduleInfo
	}

	resp := binding_constructors.NewGetModulesResponse(moduleInfos)
	return resp, nil
}

func (apicService ApiContainerService) UploadFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UploadFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse, error) {
	reader := bytes.NewReader(args.Data)

	filesArtifactUuid, err := apicService.filesArtifactStore.StoreFile(reader)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to store files.")
	}

	response := &kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse{Uuid: string(filesArtifactUuid)}
	return response, nil
}

func (apicService ApiContainerService) DownloadFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.DownloadFilesArtifactResponse, error) {
	filesArtifactUuidStr := args.Id
	if strings.TrimSpace(filesArtifactUuidStr) == "" {
		return nil, stacktrace.NewError("Cannot download file with empty files artifact UUID")
	}
	filesArtifactUuid := enclave_data_directory.FilesArtifactUUID(filesArtifactUuidStr)

	artifactFile, err := apicService.filesArtifactStore.GetFile(filesArtifactUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting files artifact '%v'", filesArtifactUuid)
	}

	fileBytes, err := ioutil.ReadFile(artifactFile.GetAbsoluteFilepath())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading files artifact file bytes")
	}

	resp := &kurtosis_core_rpc_api_bindings.DownloadFilesArtifactResponse{Data: fileBytes}
	return resp, nil
}

func (apicService ApiContainerService) StoreWebFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse, error) {
	url := args.Url

	resp, err := http.Get(args.Url)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred making the request to URL '%v' to get the files artifact bytes", url)
	}
	defer resp.Body.Close()
	body := bufio.NewReader(resp.Body)

	filesArtifactUuId, err := apicService.filesArtifactStore.StoreFile(body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred storing the file from URL '%v' in the files artifact store", url)
	}

	response := &kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse{Uuid: string(filesArtifactUuId)}
	return response, nil
}

func (apicService ApiContainerService) StoreFilesArtifactFromService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs) (*kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse, error) {
	serviceIdStr := args.ServiceId
	serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
	srcPath := args.SourcePath

	filesArtifactId, err := apicService.serviceNetwork.CopyFilesFromService(ctx, serviceId, srcPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred copying source '%v' from service with ID '%v'", srcPath, serviceId)
	}

	response := &kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse{Uuid: string(filesArtifactId)}
	return response, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func convertAPIPortsToPortSpecs(
	privateAPIPorts map[string]*kurtosis_core_rpc_api_bindings.Port,
	publicAPIPorts map[string]*kurtosis_core_rpc_api_bindings.Port,
) (
	resultPrivatePortSpecs map[string]*port_spec.PortSpec,
	resultPublicPortSpecs map[string]*port_spec.PortSpec,
	resultErr error,
) {
	privatePortSpecs := map[string]*port_spec.PortSpec{}
	for portID, privateAPIPort := range privateAPIPorts {
		privatePortSpec, err := transformApiPortToPortSpec(privateAPIPort)
		if err != nil {
			return nil, nil, stacktrace.NewError("An error occurred transforming the API port for private port '%v' into a port spec port", portID)
		}
		privatePortSpecs[portID] = privatePortSpec
	}

	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	if len(publicAPIPorts) > 0 {
		err := checkPrivateAndPublicPortsAreOneToOne(privateAPIPorts, publicAPIPorts)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Provided public and private ports are not one to one.")
		}
	}

	publicPortSpecs := map[string]*port_spec.PortSpec{}
	for portID, publicAPIPort := range publicAPIPorts {
		publicPortSpec, err := transformApiPortToPortSpec(publicAPIPort)
		if err != nil {
			return nil, nil, stacktrace.NewError("An error occurred transforming the API port for public port '%v' into a port spec port", portID)
		}
		publicPortSpecs[portID] = publicPortSpec
	}
	//TODO Finished the huge hack to temporarily enable static ports for NEAR
	return privatePortSpecs, publicPortSpecs, nil
}

func transformApiPortToPortSpec(port *kurtosis_core_rpc_api_bindings.Port) (*port_spec.PortSpec, error) {
	portNumUint32 := port.GetNumber()
	apiProto := port.GetProtocol()
	if portNumUint32 > math.MaxUint16 {
		return nil, stacktrace.NewError(
			"API port num '%v' is bigger than max allowed port spec port num '%v'",
			portNumUint32,
			math.MaxUint16,
		)
	}
	portNumUint16 := uint16(portNumUint32)
	portSpecProto, found := apiContainerPortProtoToPortSpecPortProto[apiProto]
	if !found {
		return nil, stacktrace.NewError("Couldn't find a port spec proto for API port proto '%v'; this should never happen, and is a bug in Kurtosis!", apiProto.String())
	}
	result, err := port_spec.NewPortSpec(portNumUint16, portSpecProto)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating port spec object with port num '%v' and protocol '%v'",
			portNumUint16,
			portSpecProto,
		)
	}
	return result, nil
}

func transformPortSpecToApiPort(port *port_spec.PortSpec) (*kurtosis_core_rpc_api_bindings.Port, error) {
	portNumUint16 := port.GetNumber()
	portSpecProto := port.GetProtocol()
	// Yes, this isn't the most efficient way to do this, but the map is tiny so it doesn't matter
	var apiProto kurtosis_core_rpc_api_bindings.Port_Protocol
	foundApiProto := false
	for mappedApiProto, mappedPortSpecProto := range apiContainerPortProtoToPortSpecPortProto {
		if portSpecProto == mappedPortSpecProto {
			apiProto = mappedApiProto
			foundApiProto = true
			break
		}
	}
	if !foundApiProto {
		return nil, stacktrace.NewError("Couldn't find an API port proto for port spec port proto '%v'; this should never happen, and is a bug in Kurtosis!", portSpecProto)
	}
	result := binding_constructors.NewPort(uint32(portNumUint16), apiProto)
	return result, nil
}

func transformPortSpecMapToApiPortsMap(apiPorts map[string]*port_spec.PortSpec) (map[string]*kurtosis_core_rpc_api_bindings.Port, error) {
	result := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	for portId, portSpec := range apiPorts {
		publicApiPort, err := transformPortSpecToApiPort(portSpec)
		if err != nil {
			return nil, stacktrace.NewError("An error occurred transforming port spec for port '%v' into an API port", portId)
		}
		result[portId] = publicApiPort
	}
	return result, nil
}

// Ensure that provided [privatePorts] and [publicPorts] are one to one by checking:
// - There is a matching publicPort for every portID in privatePorts
// - There are the same amount of private and public ports
// If error is nil, the public and private ports are one to one.
func checkPrivateAndPublicPortsAreOneToOne(privatePorts map[string]*kurtosis_core_rpc_api_bindings.Port, publicPorts map[string]*kurtosis_core_rpc_api_bindings.Port) error {
	if len(privatePorts) != len(publicPorts) {
		return stacktrace.NewError("The received private ports length and the public ports length are not equal. Received '%v' private ports and '%v' public ports", len(privatePorts), len(publicPorts))
	}

	for portID, privatePortSpec := range privatePorts {
		if _, found := publicPorts[portID]; !found {
			return stacktrace.NewError("Expected to receive public port with ID '%v' bound to private port number '%v', but it was not found", portID, privatePortSpec.GetNumber())
		}
	}
	return nil
}

func (apicService ApiContainerService) waitForEndpointAvailability(
	ctx context.Context,
	serviceIdStr string,
	httpMethod string,
	port uint32,
	path string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
	requestBody string,
	bodyText string) error {

	var (
		resp *http.Response
		err  error
	)

	serviceObj, err := apicService.serviceNetwork.GetService(
		ctx,
		kurtosis_backend_service.ServiceID(serviceIdStr),
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service '%v'", serviceIdStr)
	}
	if serviceObj.GetStatus() != container_status.ContainerStatus_Running {
		return stacktrace.NewError("Service '%v' isn't running so can never become available", serviceIdStr)
	}
	privateIp := serviceObj.GetRegistration().GetPrivateIP()

	url := fmt.Sprintf("http://%v:%v/%v", privateIp.String(), port, path)

	time.Sleep(time.Duration(initialDelayMilliseconds) * time.Millisecond)

	for i := uint32(0); i < retries; i++ {
		resp, err = makeHttpRequest(httpMethod, url, requestBody)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(
			err,
			"The HTTP endpoint '%v' didn't return a success code, even after %v retries with %v milliseconds in between retries",
			url,
			retries,
			retriesDelayMilliseconds,
		)
	}

	if bodyText != "" {
		body := resp.Body
		defer body.Close()

		bodyBytes, err := ioutil.ReadAll(body)

		if err != nil {
			return stacktrace.Propagate(err,
				"An error occurred reading the response body from endpoint '%v'", url)
		}

		bodyStr := string(bodyBytes)

		if bodyStr != bodyText {
			return stacktrace.NewError("Expected response body text '%v' from endpoint '%v' but got '%v' instead", bodyText, url, bodyStr)
		}
	}

	return nil
}

func makeHttpRequest(httpMethod string, url string, body string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	if httpMethod == http.MethodPost {
		var bodyByte = []byte(body)
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(bodyByte))
	} else if httpMethod == http.MethodGet {
		resp, err = http.Get(url)
	} else {
		return nil, stacktrace.NewError("HTTP method '%v' not allowed", httpMethod)
	}

	if err != nil {
		return nil, stacktrace.Propagate(err, "An HTTP error occurred when sending GET request to endpoint '%v'", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, stacktrace.NewError("Received non-OK status code: '%v'", resp.StatusCode)
	}
	return resp, nil
}

func (apicService ApiContainerService) getServiceInfo(ctx context.Context, serviceId kurtosis_backend_service.ServiceID) (*kurtosis_core_rpc_api_bindings.ServiceInfo, error) {
	serviceObj, err := apicService.serviceNetwork.GetService(
		ctx,
		serviceId,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for service '%v'", serviceId)
	}
	privatePorts := serviceObj.GetPrivatePorts()
	privateIp := serviceObj.GetRegistration().GetPrivateIP()
	maybePublicIp := serviceObj.GetMaybePublicIP()
	maybePublicPorts := serviceObj.GetMaybePublicPorts()
	serviceGuidStr := string(serviceObj.GetRegistration().GetGUID())

	privateApiPorts, err := transformPortSpecMapToApiPortsMap(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the service's private port specs to API ports")
	}
	publicIpAddrStr := missingPublicIpAddrStr
	if maybePublicIp != nil {
		publicIpAddrStr = maybePublicIp.String()
	}
	publicApiPorts := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	if maybePublicPorts != nil {
		publicApiPorts, err = transformPortSpecMapToApiPortsMap(maybePublicPorts)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred transforming the service's public port spec ports to API ports")
		}
	}

	serviceInfoResponse := binding_constructors.NewServiceInfo(
		serviceGuidStr,
		privateIp.String(),
		privateApiPorts,
		publicIpAddrStr,
		publicApiPorts,
	)
	return serviceInfoResponse, nil
}

func (apicService ApiContainerService) getModuleInfo(ctx context.Context, moduleId module.ModuleID) (*kurtosis_core_rpc_api_bindings.ModuleInfo, error) {
	moduleGuid, privateIpAddr, privateModulePort, maybePublicIpAddr, maybePublicModulePort, err := apicService.moduleStore.GetModuleInfo(moduleId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the IP address for module '%v'", moduleId)
	}
	privateApiPort, err := transformPortSpecToApiPort(privateModulePort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the module's private port spec port to an API port")
	}
	var publicApiPort *kurtosis_core_rpc_api_bindings.Port
	if maybePublicModulePort != nil {
		publicApiPort, err = transformPortSpecToApiPort(maybePublicModulePort)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred transforming the module's public port spec port to an API port")
		}
	}
	publicIpAddr := missingPublicIpAddrStr
	if maybePublicIpAddr != nil {
		publicIpAddr = maybePublicIpAddr.String()
	}
	if privateIpAddr == nil {
		return nil, stacktrace.NewError("Private IP address for module '%v' was nil - this should never happen.", moduleId)
	}
	response := binding_constructors.NewModuleInfo(
		string(moduleGuid),
		privateIpAddr.String(),
		privateApiPort,
		publicIpAddr,
		publicApiPort,
	)
	return response, nil
}