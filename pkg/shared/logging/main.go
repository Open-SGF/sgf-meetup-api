package logging

import (
	"github.com/getsentry/sentry-go"
	"log/slog"
	"os"
)

type Config struct {
	Level slog.Level
	Type  LogType
}

func DefaultLogger(config Config) *slog.Logger {
	var handler slog.Handler

	switch config.Type {
	case LogTypeText:
		handler = NewTextHandler(config.Level)
	case LogTypeJSON:
		handler = NewJSONHandler(config.Level)
	default:
		panic("unknown log type")
	}

	if sentry.CurrentHub().Client() != nil {
		handler = WithErrorLogger(handler, NewSentryErrorLogger())
	}

	return NewLogger(handler)
}

func NewLogger(handler slog.Handler) *slog.Logger {
	return slog.New(handler)
}

func NewLoggerWithErrorHandler(handler slog.Handler, errorLogger ErrorLogger) *slog.Logger {
	handlerWithErrorLogger := WithErrorLogger(handler, errorLogger)

	return NewLogger(handlerWithErrorLogger)
}

func NewJSONHandler(level slog.Level) *slog.JSONHandler {
	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
}

func NewTextHandler(level slog.Level) *slog.TextHandler {
	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
}
