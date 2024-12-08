package resources

import (
	"context"
	"errors"
)

type Resource interface {
	Id() string
	WatchChecker
	Dependant
}

type WatchChecker interface {
	Watcher
	Checker
}

type Correction func() error

type Checker interface {
	Check() ([]Correction, error)
}

type Watcher interface {
	Watch(ctx context.Context, correctionsCh chan<- []Correction, errCh chan<- error)
}

type Dependant interface {
	Dependencies() []Resource
	SetDependencies(dependencies ...Resource)
}

var ErrUnalignedResource = errors.New("unaligned resource")

type BaseDependant struct {
	dependencies []Resource
}

func (d *BaseDependant) Dependencies() []Resource {
	return d.dependencies
}

func (d *BaseDependant) SetDependencies(dependencies ...Resource) {
	d.dependencies = dependencies
}
