package resources

import (
	"context"
	"errors"
)

type Correction func() error

type ResourceWatcher interface {
	Resource
	Watcher
}

type Resource interface {
	Id() string
	Check() ([]Correction, error)
}

type Watcher interface {
	Watch(ctx context.Context, correctionsCh chan<- []Correction, errCh chan<- error)
}

var ErrUnalignedResource = errors.New("unaligned resource")
