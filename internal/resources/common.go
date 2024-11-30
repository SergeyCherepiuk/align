package resources

import (
	"errors"

	"github.com/SergeyCherepiuk/align/internal/logger"
)

func checkAndCorrect(resource Resource) error {
	logger.Global().Info("checking resource", "resource", resource.Id())

	corrections, err := resource.Check()

	if errors.Is(err, ErrUnalignedResource) {
		logger.Global().Info("executing corrections", "resource", resource.Id(), "count", len(corrections))

		err := executeCorrections(corrections)
		if err != nil {
			logger.Global().Error("correction failed", "resource", resource.Id(), "error", err.Error())
			return err
		}

		logger.Global().Info("corrections executed succesfully", "resource", resource.Id())

		return nil
	}

	if err != nil {
		return err
	}

	logger.Global().Info("resource is aligned", "resource", resource.Id())

	return nil
}

func executeCorrections(corrections []Correction) error {
	for _, correction := range corrections {
		err := correction()
		if err != nil {
			return err
		}
	}

	return nil
}
