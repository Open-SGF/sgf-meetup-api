package logging

import (
	"log/slog"
	"os"
)

func DefaultLogger(level slog.Level) *slog.Logger {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	handler := WithErrorLogger(baseHandler, &SentryErrorLogger{})

	return slog.New(handler)
}
