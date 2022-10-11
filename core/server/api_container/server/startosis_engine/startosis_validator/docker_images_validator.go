package startosis_validator

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

type DockerImagesValidator struct {
	kurtosisBackend *backend_interface.KurtosisBackend
}

func NewDockerImagesValidator(kurtosisBackend *backend_interface.KurtosisBackend) *DockerImagesValidator {
	return &DockerImagesValidator{
		kurtosisBackend,
	}
}

func (validator *DockerImagesValidator) Validate(ctx context.Context, environment *ValidatorEnvironment) []error {
	pullErrors := make(chan error, len(environment.requiredDockerImages))
	var wg sync.WaitGroup
	for image := range environment.requiredDockerImages {
		wg.Add(1)
		go fetchImageFromBackend(ctx, &wg, validator.kurtosisBackend, image, pullErrors)
	}
	wg.Wait()
	close(pullErrors)
	if len(pullErrors) > 0 {
		errors := []error{}
		for pullError := range pullErrors {
			errors = append(errors, pullError)
		}
		return errors
	}
	return nil
}

func fetchImageFromBackend(ctx context.Context, wg *sync.WaitGroup, backend *backend_interface.KurtosisBackend, image string, pullError chan<- error) {
	defer wg.Done()
	err := (*backend).FetchImage(ctx, image)
	if err != nil {
		pullError <- stacktrace.Propagate(err, "Failed fetching the required image '%v'", image)
	}
}
