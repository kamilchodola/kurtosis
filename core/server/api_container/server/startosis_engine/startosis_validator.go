package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
)

type StartosisValidator struct {
	dockerImagesValidator *startosis_validator.DockerImagesValidator
}

func NewStartosisValidator(kurtosisBackend *backend_interface.KurtosisBackend) *StartosisValidator {
	dockerImagesValidator := startosis_validator.NewDockerImagesValidator(kurtosisBackend)
	return &StartosisValidator{
		dockerImagesValidator,
	}
}

func (validator *StartosisValidator) Validate(ctx context.Context, instructions []kurtosis_instruction.KurtosisInstruction) error {
	environment := startosis_validator.NewValidatorEnvironment()
	for _, instruction := range instructions {
		err := instruction.ValidateAndUpdateEnvironment(environment)
		if err != nil {
			return stacktrace.Propagate(err, "Error while validating instruction %v", instruction.String())
		}
	}
	err := validator.dockerImagesValidator.Validate(ctx, environment)
	if err != nil {
		return stacktrace.Propagate(err, "Error while validating final environment of script")
	}
	return nil
}
