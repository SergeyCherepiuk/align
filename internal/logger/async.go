package logger

import (
	"context"
	"log/slog"
	"sync"
)

const (
	workers    = 5
	bufferSize = 512
)

type log struct {
	level slog.Level
	msg   string
	args  []any
}

type asyncLogger struct {
	ch chan log
	wg sync.WaitGroup
}

func newAsyncLogger(ctx context.Context) *asyncLogger {
	// TODO: sc: Move log-level setting to the interface level.
	slog.SetLogLoggerLevel(slog.LevelDebug)

	logger := asyncLogger{ch: make(chan log, bufferSize)}

	logger.wg.Add(workers)
	for range workers {
		go logger.worker(ctx)
	}

	return &logger
}

func (l *asyncLogger) Debug(msg string, args ...any) {
	l.ch <- log{slog.LevelDebug, msg, args}
}

func (l *asyncLogger) Info(msg string, args ...any) {
	l.ch <- log{slog.LevelInfo, msg, args}
}

func (l *asyncLogger) Warn(msg string, args ...any) {
	l.ch <- log{slog.LevelWarn, msg, args}
}

func (l *asyncLogger) Error(msg string, args ...any) {
	l.ch <- log{slog.LevelError, msg, args}
}

func (l *asyncLogger) Close() error {
	close(l.ch)
	l.wg.Wait()
	return nil
}

func (l *asyncLogger) worker(ctx context.Context) {
	defer l.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case log, ok := <-l.ch:
			if !ok {
				return
			}

			slog.Log(ctx, log.level, log.msg, log.args...)
		}
	}
}
