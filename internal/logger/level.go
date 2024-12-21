package logger

import "log/slog"

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type slogLeveler struct {
	level slog.Level
}

func newSlogLeveler(level Level) *slogLeveler {
	leveler := new(slogLeveler)
	leveler.setLevel(level)
	return leveler
}

func (l *slogLeveler) Level() slog.Level {
	return l.level
}

func (l *slogLeveler) setLevel(level Level) {
	switch level {
	case LevelDebug:
		l.level = slog.LevelDebug
	case LevelInfo:
		l.level = slog.LevelInfo
	case LevelWarn:
		l.level = slog.LevelWarn
	case LevelError:
		l.level = slog.LevelError
	}
}
