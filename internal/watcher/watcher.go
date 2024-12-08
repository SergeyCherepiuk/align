package watcher

import (
	"context"

	"github.com/SergeyCherepiuk/align/internal/resources"
)

type ResourceWatcher struct {
	resources []resources.Resource
}

// TODO: sc: Traverse the dependency graph looking for cycles.
func NewResourceWatcher(resources ...resources.Resource) (*ResourceWatcher, error) {
	return &ResourceWatcher{resources}, nil
}

// TODO: sc: Start watching resources accounting for dependecies.
func (w *ResourceWatcher) Watch(ctx context.Context) error {
	correctionsCh := make(chan []resources.Correction)
	errCh := make(chan error)

	for _, resource := range w.resources {
		go resource.Watch(ctx, correctionsCh, errCh)
	}

	return executeCorrections(correctionsCh, errCh)
}

func executeCorrections(correctionsCh <-chan []resources.Correction, errCh <-chan error) error {
	for {
		select {
		case corrections := <-correctionsCh:
			for _, correction := range corrections {
				err := correction()
				if err != nil {
					return err
				}
			}

		case err := <-errCh:
			return err
		}
	}
}
