package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules/mock_module_manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var testServiceNetwork *service_network.ServiceNetwork = nil

func TestStartosisInterpreter_SimplePrintScript(t *testing.T) {
	testString := "Hello World!"
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	startosisInterpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	interpreter := startosisInterpreter
	script := `
print("` + testString + `")
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions)) // No kurtosis instruction
	require.Nil(t, interpretationError)

	expectedOutput := testString + `
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_ScriptFailingSingleError(t *testing.T) {
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
print("Starting Startosis script!")

unknownInstruction()
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		"Multiple errors caught interpreting the Startosis script. Listing each of them below.",
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(4, 1)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ScriptFailingMultipleErrors(t *testing.T) {
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
print("Starting Startosis script!")

unknownInstruction()
print(unknownVariable)

unknownInstruction2()
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		multipleInterpretationErrorMsg,
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(4, 1)),
			*startosis_errors.NewCallFrame("undefined: unknownVariable", startosis_errors.NewScriptPosition(5, 7)),
			*startosis_errors.NewCallFrame("undefined: unknownInstruction2", startosis_errors.NewScriptPosition(7, 1)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ScriptFailingSyntaxError(t *testing.T) {
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
print("Starting Startosis script!")

load("otherScript.start") # fails b/c load takes in at least 2 args
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)

	expectedError := startosis_errors.NewInterpretationErrorFromStacktrace(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("load statement must import at least 1 symbol", startosis_errors.NewScriptPosition(4, 5)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstruction(t *testing.T) {
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 1, len(instructions))
	require.Nil(t, interpretationError)

	addServiceInstruction := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(13, 12),
		service.ServiceID("example-datastore-server"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})

	require.Equal(t, instructions[0], addServiceInstruction)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionMissingContainerName(t *testing.T) {
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	# /!\ /!\ missing container name /!\ /!\
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		"Evaluation error: Missing value 'container_image_name' as element of the struct object 'service_config'",
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(13, 12)),
			*startosis_errors.NewCallFrame("add_service", startosis_errors.NewScriptPosition(0, 0)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionTypoInProtocol(t *testing.T) {
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCPK") # typo in protocol
	}
)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		"Evaluation error: Port protocol should be one of TCP, SCTP, UDP",
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(13, 12)),
			*startosis_errors.NewCallFrame("add_service", startosis_errors.NewScriptPosition(0, 0)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionPortNumberAsString(t *testing.T) {
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = "1234", protocol = "TCP") # port number should be an int
	}
)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		"Evaluation error: Argument 'number' is expected to be an integer. Got starlark.String",
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(13, 12)),
			*startosis_errors.NewCallFrame("add_service", startosis_errors.NewScriptPosition(0, 0)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidScriptWithMultipleInstructions(t *testing.T) {
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
ports = [1323, 1324, 1325]

def deploy_datastore_services():
    for i in range(len(ports)):
        unique_service_id = service_id + "-" + str(i)
        print("Adding service " + unique_service_id)
        service_config = struct(
			container_image_name = "kurtosistech/example-datastore-server",
			used_ports = {
				"grpc": struct(
					number = ports[i],
					protocol = "TCP"
				)
			}
		)
        add_service(service_id = unique_service_id, service_config = service_config)

deploy_datastore_services()
print("Done!")
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 3, len(instructions))
	require.Nil(t, interpretationError)

	addServiceInstruction0 := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(22, 26),
		service.ServiceID("example-datastore-server-0"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})
	addServiceInstruction1 := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(22, 26),
		service.ServiceID("example-datastore-server-1"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1324,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})
	addServiceInstruction2 := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(22, 26),
		service.ServiceID("example-datastore-server-2"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1325,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})

	require.Equal(t, instructions[0], addServiceInstruction0)
	require.Equal(t, instructions[1], addServiceInstruction1)
	require.Equal(t, instructions[2], addServiceInstruction2)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_SimpleLoading(t *testing.T) {
	barModulePath := "github.com/foo/bar/lib.star"
	seedModules := map[string]string {
		barModulePath: "a=\"World!\"",
	}
	moduleManager := mock_module_manager.NewMockModuleManager(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
load("` + barModulePath + `", "a")
print("Hello " + a)
`
	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction
	assert.Nil(t, interpretationError)

	expectedOutput := `Hello World!
`
	assert.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_TransitiveLoading(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `a="World!"`
	moduleDooWhichLoadsModuleBar := "github.com/foo/doo/lib.star"
	seedModules[moduleDooWhichLoadsModuleBar] = `load("` + moduleBar + `", "a")
b = "Hello " + a
`
	moduleManager := mock_module_manager.NewMockModuleManager(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
load("` + moduleDooWhichLoadsModuleBar + `", "b")
print(b)

`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction
	assert.Nil(t, interpretationError)

	expectedOutput := `Hello World!
`
	assert.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_FailsOnCycle(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBarLoadsModuleDoo := "github.com/foo/bar/lib.star"
	moduleDooLoadsModuleBar := "github.com/foo/doo/lib.star"
	seedModules[moduleBarLoadsModuleDoo] = `load("` + moduleDooLoadsModuleBar + `", "b")
a = "Hello" + b`
	seedModules[moduleDooLoadsModuleBar] = `load("` + moduleBarLoadsModuleDoo + `", "a")
b = "Hello " + a
`
	moduleManager := mock_module_manager.NewMockModuleManager(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
load("` + moduleDooLoadsModuleBar + `", "b")
print(b)
`

	_, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		fmt.Sprintf("Evaluation error: cannot load %v: cannot load %v: cannot load %v: There is a cycle in the load graph", moduleDooLoadsModuleBar, moduleBarLoadsModuleDoo, moduleDooLoadsModuleBar),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(2, 1)),
		},
	)
	assert.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_FailsOnNonExistentModule(t *testing.T) {
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	nonExistentModule := "github.com/non/existent/module.star"
	script := `
load("` + nonExistentModule + `", "b")
print(b)
`
	_, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		fmt.Sprintf("Evaluation error: cannot load %v: An error occurred while loading the module '%v'", nonExistentModule, nonExistentModule),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(2, 1)),
		},
	)
	assert.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_LoadingAValidModuleThatPreviouslyFailedToLoadSucceeds(t *testing.T) {
	moduleManager := mock_module_manager.NewEmptyMockModuleManager()
	barModulePath := "github.com/foo/bar/lib.star"
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
load("` + barModulePath + `", "a")
print("Hello " + a)
`

	// assert that first load fails
	_, interpretationError, _ := interpreter.Interpret(context.Background(), script)
	assert.NotNil(t, interpretationError)


	barModuleContents := "a=\"World!\""
	moduleManager.Add(barModulePath, barModuleContents)
	expectedOutput := `Hello World!
`
	// assert that second load succeeds
	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Nil(t, interpretationError)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction
	assert.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_ValidSimpleScriptWithImportedStruct(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
print("Constructing service_config")
service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
`
	moduleManager := mock_module_manager.NewMockModuleManager(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
load("` + moduleBar + `", "service_id", "service_config")
print("Starting Startosis script!")

print("Adding service " + service_id)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 1, len(instructions))
	require.Nil(t, interpretationError)

	addServiceInstruction := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(6, 12),
		service.ServiceID("example-datastore-server"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})

	require.Equal(t, instructions[0], addServiceInstruction)

	expectedOutput := `Constructing service_config
Starting Startosis script!
Adding service example-datastore-server
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_ValidScriptWithFunctionsImportedFromOtherModule(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
ports = [1323, 1324, 1325]

def deploy_datastore_services():
    for i in range(len(ports)):
        unique_service_id = service_id + "-" + str(i)
        print("Adding service " + unique_service_id)
        service_config = struct(
			container_image_name = "kurtosistech/example-datastore-server",
			used_ports = {
				"grpc": struct(
					number = ports[i],
					protocol = "TCP"
				)
			}
		)
        add_service(service_id = unique_service_id, service_config = service_config)
`
	moduleManager := mock_module_manager.NewMockModuleManager(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
load("` + moduleBar + `", "deploy_datastore_services")
print("Starting Startosis script!")

deploy_datastore_services()
print("Done!")
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 3, len(instructions))
	require.Nil(t, interpretationError)

	addServiceInstruction0 := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(5, 26),
		service.ServiceID("example-datastore-server-0"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})
	addServiceInstruction1 := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(5, 26),
		service.ServiceID("example-datastore-server-1"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1324,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})
	addServiceInstruction2 := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(5, 26),
		service.ServiceID("example-datastore-server-2"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1325,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})

	require.Equal(t, instructions[0], addServiceInstruction0)
	require.Equal(t, instructions[1], addServiceInstruction1)
	require.Equal(t, instructions[2], addServiceInstruction2)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_AddServiceInOtherModulePopulatesQueue(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
print("Constructing service_config")
service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
print("Adding service " + service_id)
add_service(service_id = service_id, service_config = service_config)
`
	moduleManager := mock_module_manager.NewMockModuleManager(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	script := `
load("` + moduleBar + `", "service_id", "service_config")
print("Starting Startosis script!")
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 1, len(instructions))
	require.Nil(t, interpretationError)

	addServiceInstruction := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(11, 12),
		service.ServiceID("example-datastore-server"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})

	require.Equal(t, instructions[0], addServiceInstruction)

	expectedOutput := `Constructing service_config
Adding service example-datastore-server
Starting Startosis script!
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_TestInstructionQueueAndOutputBufferDontHaveDupesInterpretingAnotherScript(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
print("Constructing service_config")
service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
print("Adding service " + service_id)
add_service(service_id = service_id, service_config = service_config)
`
	moduleManager := mock_module_manager.NewMockModuleManager(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleManager)
	scriptA := `
load("` + moduleBar + `", "service_id", "service_config")
print("Starting Startosis script!")
`
	addServiceInstructionFromScriptA := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(11, 12),
		service.ServiceID("example-datastore-server"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})

	expectedOutputFromScriptA := `Constructing service_config
Adding service example-datastore-server
Starting Startosis script!
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), scriptA)
	require.Equal(t, 1, len(instructions))
	require.Nil(t, interpretationError)
	require.Equal(t, instructions[0], addServiceInstructionFromScriptA)
	require.Equal(t, expectedOutputFromScriptA, string(scriptOutput))

	scriptB := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
add_service(service_id = service_id, service_config = service_config)
`
	addServiceInstructionFromScriptB := add_service.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(13, 12),
		service.ServiceID("example-datastore-server"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
	})
	expectedOutputFromScriptB := `Starting Startosis script!
Adding service example-datastore-server
`

	scriptOutput, interpretationError, instructions = interpreter.Interpret(context.Background(), scriptB)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, len(instructions))
	require.Equal(t, instructions[0], addServiceInstructionFromScriptB)
	require.Equal(t, expectedOutputFromScriptB, string(scriptOutput))
}

