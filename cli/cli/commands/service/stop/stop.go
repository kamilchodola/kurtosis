package stop

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/service_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceIdentifierArgKey        = "service"
	isServiceIdentifierArgOptional = false
	isServiceIdentifierArgGreedy   = true

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	starlarkScript = `
def run(plan, args):
	plan.stop_service(name=args["service_name"])
`
)

var ServiceStopCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceStopCmdStr,
	ShortDescription:          "Stops a service",
	LongDescription:           "Stops temporarily a service with the given service identifier in the given enclave",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		service_identifier_arg.NewServiceIdentifierArg(
			serviceIdentifierArgKey,
			enclaveIdentifierArgKey,
			isServiceIdentifierArgOptional,
			isServiceIdentifierArgGreedy,
		),
	},
	Flags:   []*flags.FlagConfig{},
	RunFunc: run,
}

func run(
	ctx context.Context,
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	_ *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifier value using key '%v'", enclaveIdentifierArgKey)
	}

	serviceIdentifiers, err := args.GetGreedyArg(serviceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service identifier value using key '%v'", serviceIdentifierArgKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveIdentifier)
	}

	for _, serviceIdentifier := range serviceIdentifiers {
		logrus.Infof("Stopping service '%v'", serviceIdentifier)
		serviceContext, err := enclaveCtx.GetServiceContext(serviceIdentifier)
		if err != nil {
			return stacktrace.NewError("Couldn't validate whether the service exists for identifier '%v'", serviceIdentifier)
		}

		serviceName := serviceContext.GetServiceName()

		if err := stopServiceStarlarkCommand(ctx, enclaveCtx, serviceName); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping service '%v' from enclave '%v'", serviceIdentifier, enclaveIdentifier)
		}
	}
	return nil
}

func stopServiceStarlarkCommand(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, serviceName services.ServiceName) error {
	params := fmt.Sprintf(`{"service_name": "%s"}`, serviceName)
	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams(params)))
	if err != nil {
		return stacktrace.Propagate(err, "An unexpected error occurred on Starlark for stopping service")
	}
	if runResult.ExecutionError != nil {
		return stacktrace.NewError("An error occurred during Starlark script execution for stopping service: %s", runResult.ExecutionError.GetErrorMessage())
	}
	if runResult.InterpretationError != nil {
		return stacktrace.NewError("An error occurred during Starlark script interpretation for stopping service: %s", runResult.InterpretationError.GetErrorMessage())
	}
	if len(runResult.ValidationErrors) > 0 {
		return stacktrace.NewError("An error occurred during Starlark script validation for stopping service: %v", runResult.ValidationErrors)
	}
	return nil
}
