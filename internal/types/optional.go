package types

type Optional[T any] struct {
	value T
	ok    bool
}

func NewOptional[T any](value T) Optional[T] {
	return Optional[T]{value, true}
}

func (o Optional[T]) Value() T {
	return o.value
}

func (o Optional[T]) Ok() bool {
	return o.ok
}
