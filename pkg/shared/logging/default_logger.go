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

	opts := &slog.HandlerOptions{
		Level: config.Level,
	}

	switch config.Type {
	case LogTypeText:
		handler = slog.NewTextHandler(os.Stdout, opts)
	case LogTypeJSON:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		panic("unknown log type")
	}

	if sentry.CurrentHub().Client() != nil {
		handler = WithErrorLogger(handler, NewSentryErrorLogger())
	}

	return slog.New(handler)
}
