package logger

import (
	"context"
	"sync"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Close() error
}

var (
	_logger Logger = &noopLogger{}
	once    sync.Once
)

func Setup(ctx context.Context) {
	setup := func() { _logger = newAsyncLogger(ctx) }
	once.Do(setup)
}

func Global() Logger {
	return _logger
}
