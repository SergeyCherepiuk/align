package watcher

import (
	"context"
	"errors"
	"fmt"

	"github.com/scherepiuk/align/internal/resources"
)

type resourceWatcher struct {
	dependencyLayers [][]resources.Resource
}

func NewResourceWatcher(resources ...resources.Resource) (*resourceWatcher, error) {
	layers, err := sortTopologically(resources)
	if err != nil {
		return nil, fmt.Errorf("failed to construct dependency graph: %w", err)
	}

	return &resourceWatcher{layers}, nil
}

func (w *resourceWatcher) Watch(ctx context.Context) error {
	correctionsCh := make(chan []resources.Correction)
	errCh := make(chan error)

	for _, layer := range w.dependencyLayers {
		for _, resource := range layer {
			go func() { errCh <- checkAndExecuteCorrections(resource) }()
		}

		for range len(layer) {
			err := <-errCh
			if err != nil {
				return err
			}
		}
	}

	for _, layer := range w.dependencyLayers {
		for _, resource := range layer {
			go resource.Watch(ctx, correctionsCh, errCh)
		}
	}

	return startExecutingCorrections(ctx, correctionsCh, errCh)
}

func checkAndExecuteCorrections(resource resources.Resource) error {
	corrections, err := resource.Check()

	if errors.Is(err, resources.ErrUnalignedResource) {
		return executeCorrections(corrections)
	}

	return err
}

func startExecutingCorrections(
	ctx context.Context,
	correctionsCh <-chan []resources.Correction,
	errCh <-chan error,
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case corrections := <-correctionsCh:
			err := executeCorrections(corrections)
			if err != nil {
				return err
			}

		case err := <-errCh:
			return err
		}
	}
}

func executeCorrections(corrections []resources.Correction) error {
	for _, correction := range corrections {
		err := correction()
		if err != nil {
			return fmt.Errorf("failed to execute corrections: %w", err)
		}
	}

	return nil
}
