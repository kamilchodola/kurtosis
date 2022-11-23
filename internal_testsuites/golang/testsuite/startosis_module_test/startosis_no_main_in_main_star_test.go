package startosis_module_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	invalidCaseNoMainInMainStarTestName = "invalid-module-missing-main"
	moduleWithNoMainInMainStarRelPath   = "../../../startosis/no-main-in-main-star"
)

func TestStartosisModule_NoMainInMainStar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidCaseNoMainInMainStarTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, moduleWithNoMainInMainStarRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	expectedInterpretationErr := "Evaluation error: module has no .main field or method\n\tat [3:12]: <toplevel>"
	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	require.NotNil(t, executionResult.GetInterpretationError())
	require.Contains(t, executionResult.GetInterpretationError().GetErrorMessage(), expectedInterpretationErr)
	require.Nil(t, executionResult.GetValidationErrors())
	require.Nil(t, executionResult.GetExecutionError())
	require.Empty(t, test_helpers.GenerateScriptOutput(executionResult.GetKurtosisInstructions()))
}
