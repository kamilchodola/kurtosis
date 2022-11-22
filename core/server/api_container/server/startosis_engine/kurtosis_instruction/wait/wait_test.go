package wait

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testServiceId = "example-service-id"
	testFactName  = "example-fact-name"
)

var (
	emptyFactsEngine *facts_engine.FactsEngine = nil
)

func TestWaitInstruction_GetCanonicalizedInstruction(t *testing.T) {
	execInstruction := NewWaitInstruction(
		emptyFactsEngine,
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		testServiceId,
		testFactName,
	)
	expectedMultiLineFormatStr := `# from: dummyFile[1:1]
wait(
	fact_name="%v",
	service_id="%v"
)`
	expectedMultiLineStr := fmt.Sprintf(expectedMultiLineFormatStr, testFactName, testServiceId)
	require.Equal(t, expectedMultiLineStr, execInstruction.GetCanonicalInstruction())

	expectedSingleLineFormatStr := `wait(fact_name="%v", service_id="%v")`
	expectedSingleLineStr := fmt.Sprintf(expectedSingleLineFormatStr, testFactName, testServiceId)
	require.Equal(t, expectedSingleLineStr, execInstruction.String())
}
