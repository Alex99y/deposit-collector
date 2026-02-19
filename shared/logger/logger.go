package logger

import (
	slog "log/slog"
	"os"
)

type Logger struct {
	logger *slog.Logger
}

func NewLogger() *Logger {
	return &Logger{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func (l *Logger) Info(message string) {
	l.logger.Info(message)
}

func (l *Logger) Error(message string) {
	l.logger.Error(message)
}

func (l *Logger) Debug(message string) {
	l.logger.Debug(message)
}

func (l *Logger) Warn(message string) {
	l.logger.Warn(message)
}
