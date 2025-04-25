package infraconfig

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"sgf-meetup-api/pkg/shared/appconfig"
	"sgf-meetup-api/pkg/shared/logging"
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Run("all values", func(t *testing.T) {
		tempDir, cleanup, err := appconfig.SetupTestEnv(`
LOG_LEVEL=debug
LOG_TYPE=json
APP_ENV=staging
`)

		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(context.Background(), tempDir, ".env")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, slog.LevelDebug, cfg.LogLevel)
		assert.Equal(t, logging.LogTypeJSON, cfg.LogType)
		assert.Equal(t, "staging", cfg.AppEnv)
	})

	t.Run("minimal with defaults", func(t *testing.T) {
		tempDir, cleanup, err := appconfig.SetupTestEnv(`
`)

		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(context.Background(), tempDir, ".env")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, slog.LevelInfo, cfg.LogLevel)
		assert.Equal(t, logging.LogTypeText, cfg.LogType)
	})
}

func TestNewLoggingConfig(t *testing.T) {
	cfg := &Config{
		LogLevel: slog.LevelDebug,
		LogType:  logging.LogTypeJSON,
	}

	loggingCfg := NewLoggingConfig(cfg)
	assert.Equal(t, slog.LevelDebug, loggingCfg.Level)
	assert.Equal(t, logging.LogTypeJSON, loggingCfg.Type)
}
