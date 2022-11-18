package startosis_engine

import (
	"context"
	"errors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/mock_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	executeSuccessfully = true
	throwOnExecute      = false

	doDryRun       = true
	executeForReal = false
)

var (
	intoTheVoid = &strings.Builder{}
)

func TestExecuteKurtosisInstructions_ExecuteForReal_Success(t *testing.T) {
	executor := NewStartosisExecutor()

	instruction1 := createMockInstruction(t, "instruction1()", executeSuccessfully)
	instruction2 := createMockInstruction(t, "instruction2()", executeSuccessfully)
	instructions := []kurtosis_instruction.KurtosisInstruction{
		instruction1,
		instruction2,
	}

	serializedInstruction, err := executor.Execute(context.Background(), executeForReal, instructions, intoTheVoid)
	instruction1.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction1.AssertNumberOfCalls(t, "Execute", 1)
	instruction2.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction2.AssertNumberOfCalls(t, "Execute", 1)

	require.Nil(t, err)

	expectedSerializedInstructions := []*kurtosis_core_rpc_api_bindings.SerializedKurtosisInstruction{
		binding_constructors.NewSerializedKurtosisInstruction("instruction1()"),
		binding_constructors.NewSerializedKurtosisInstruction("instruction2()"),
	}
	require.Equal(t, serializedInstruction, expectedSerializedInstructions)
}

func TestExecuteKurtosisInstructions_ExecuteForReal_FailureHalfWay(t *testing.T) {
	executor := NewStartosisExecutor()

	instruction1 := createMockInstruction(t, "instruction1()", executeSuccessfully)
	instruction2 := createMockInstruction(t, "instruction2()", throwOnExecute)
	instruction3 := createMockInstruction(t, "instruction3()", executeSuccessfully)
	instructions := []kurtosis_instruction.KurtosisInstruction{
		instruction1,
		instruction2,
		instruction3,
	}

	serializedInstruction, err := executor.Execute(context.Background(), executeForReal, instructions, intoTheVoid)
	instruction1.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction1.AssertNumberOfCalls(t, "Execute", 1)
	instruction2.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction2.AssertNumberOfCalls(t, "Execute", 1)
	// nothing called for instruction 3 because instruction 2 threw an error
	instruction3.AssertNumberOfCalls(t, "GetCanonicalInstruction", 0)
	instruction3.AssertNumberOfCalls(t, "Execute", 0)

	expectedErrorMsgPrefix := `An error occurred executing instruction (number 2): 
instruction2()
 --- at`
	expectedLowLevelErrorMessage := "expected error for test"
	require.NotNil(t, err)
	require.Contains(t, err.Error(), expectedErrorMsgPrefix)
	require.Contains(t, err.Error(), expectedLowLevelErrorMessage)

	expectedSerializedInstructions := []*kurtosis_core_rpc_api_bindings.SerializedKurtosisInstruction{
		// only instruction 1 because it failed at instruction 2
		binding_constructors.NewSerializedKurtosisInstruction("instruction1()"),
	}
	require.Equal(t, serializedInstruction, expectedSerializedInstructions)
}

func TestExecuteKurtosisInstructions_DoDryRun(t *testing.T) {
	executor := NewStartosisExecutor()

	instruction1 := createMockInstruction(t, "instruction1()", executeSuccessfully)
	instruction2 := createMockInstruction(t, "instruction2()", executeSuccessfully)
	instructions := []kurtosis_instruction.KurtosisInstruction{
		instruction1,
		instruction2,
	}

	serializedInstruction, err := executor.Execute(context.Background(), doDryRun, instructions, intoTheVoid)
	instruction1.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction2.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	// both execute never called because dry run = true
	instruction1.AssertNumberOfCalls(t, "Execute", 0)
	instruction2.AssertNumberOfCalls(t, "Execute", 0)

	require.Nil(t, err)

	expectedSerializedInstructions := []*kurtosis_core_rpc_api_bindings.SerializedKurtosisInstruction{
		binding_constructors.NewSerializedKurtosisInstruction("instruction1()"),
		binding_constructors.NewSerializedKurtosisInstruction("instruction2()"),
	}
	require.Equal(t, serializedInstruction, expectedSerializedInstructions)
}

func createMockInstruction(t *testing.T, canonicalizedInstruction string, executeSuccessfully bool) *mock_instruction.MockKurtosisInstruction {
	instruction := mock_instruction.NewMockKurtosisInstruction(t)

	instruction.EXPECT().GetCanonicalInstruction().Maybe().Return(canonicalizedInstruction)

	if executeSuccessfully {
		instruction.EXPECT().Execute(mock.Anything).Maybe().Return(nil, nil)
	} else {
		instruction.EXPECT().Execute(mock.Anything).Maybe().Return(nil, errors.New("expected error for test"))
	}

	return instruction
}
