package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/SergeyCherepiuk/align/internal/logger"
	"github.com/SergeyCherepiuk/align/internal/resources"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Setup(ctx)
	defer logger.Global().Close()

	for _, resource := range expectedResources() {
		checkResource(resource)
	}
}

func expectedResources() []resources.Resource {
	return []resources.Resource{
		resources.NewFile(
			"/tmp/align-testing",
			resources.WithMode(os.FileMode(0o664)),
			resources.WithOwner("scherepiuk"),
			resources.WithGroup("scherepiuk"),
		),
	}
}

func checkResource(resource resources.Resource) {
	logger.Global().Info("checking resource", "resource", resource.Id())

	corrections, err := resource.Check()

	if !errors.Is(err, resources.ErrUnalignedResource) {
		logger.Global().Info("resource is aligned", "resource", resource.Id())
		return
	}

	if errors.Is(err, resources.ErrUnalignedResource) {
		logger.Global().Info("execute corrections", "resource", resource.Id(), "count", len(corrections))

		err := executeCorrections(corrections)
		if err != nil {
			log.Fatal(err)
		}

		logger.Global().Info("corrections executed succesfully", "resource", resource.Id())
		return
	}

	if err != nil {
		log.Fatal(err)
	}
}

func executeCorrections(corrections []resources.Correction) error {
	for _, correction := range corrections {
		err := correction()
		if err != nil {
			return fmt.Errorf("correction failed: %w", err)
		}
	}

	return nil
}
