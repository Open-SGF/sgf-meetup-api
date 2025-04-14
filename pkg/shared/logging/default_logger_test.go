package logging

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	t.Run("creates json handler", func(t *testing.T) {
		logger := DefaultLogger(Config{Level: slog.LevelWarn, Type: LogTypeJSON})

		assert.IsType(t, new(slog.JSONHandler), logger.Handler())
		assert.True(t, logger.Handler().Enabled(context.Background(), slog.LevelWarn))
		assert.False(t, logger.Handler().Enabled(context.Background(), slog.LevelInfo))
	})

	t.Run("adds error logger if sentry is enabled", func(t *testing.T) {
		_ = sentry.Init(sentry.ClientOptions{})
		defer sentry.CurrentHub().BindClient(nil)

		logger := DefaultLogger(Config{Level: slog.LevelWarn, Type: LogTypeJSON})

		assert.IsType(t, new(captureErrorHandler), logger.Handler())

		handler, _ := logger.Handler().(*captureErrorHandler)

		assert.IsType(t, new(SentryErrorLogger), handler.errorLogger)
		assert.IsType(t, new(slog.JSONHandler), handler.handler)
	})

	t.Run("creates text handler", func(t *testing.T) {
		logger := DefaultLogger(Config{Level: slog.LevelWarn, Type: LogTypeText})

		assert.IsType(t, new(slog.TextHandler), logger.Handler())
		assert.True(t, logger.Handler().Enabled(context.Background(), slog.LevelWarn))
		assert.False(t, logger.Handler().Enabled(context.Background(), slog.LevelInfo))
	})

	t.Run("panics for unknown log type", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = DefaultLogger(Config{Level: slog.LevelWarn, Type: LogType(-1)})
		})
	})
}
