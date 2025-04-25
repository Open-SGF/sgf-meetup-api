package syncdynamodbconfig

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
DYNAMODB_ENDPOINT=http://localhost:8000
AWS_REGION=us-west-2
AWS_ACCESS_KEY=testkey
AWS_SECRET_ACCESS_KEY=testsecret
`)

		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(context.Background(), tempDir, ".env")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, slog.LevelDebug, cfg.LogLevel)
		assert.Equal(t, logging.LogTypeJSON, cfg.LogType)
		assert.Equal(t, "http://localhost:8000", cfg.DynamoDbEndpoint)
		assert.Equal(t, "us-west-2", cfg.AwsRegion)
		assert.Equal(t, "testkey", cfg.AwsAccessKey)
		assert.Equal(t, "testsecret", cfg.AwsSecretAccessKey)
	})

	t.Run("minimal with defaults", func(t *testing.T) {
		tempDir, cleanup, err := appconfig.SetupTestEnv(``)

		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(context.Background(), tempDir, ".env")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, slog.LevelInfo, cfg.LogLevel)
		assert.Equal(t, logging.LogTypeText, cfg.LogType)
		assert.Equal(t, "us-east-2", cfg.AwsRegion)
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

func TestNewDBConfig(t *testing.T) {
	cfg := &Config{
		DynamoDbEndpoint:   "http://localhost:8000",
		AwsRegion:          "us-west-2",
		AwsAccessKey:       "testkey",
		AwsSecretAccessKey: "testsecret",
	}

	dbCfg := NewDBConfig(cfg)
	assert.Equal(t, "http://localhost:8000", dbCfg.Endpoint)
	assert.Equal(t, "us-west-2", dbCfg.Region)
	assert.Equal(t, "testkey", dbCfg.AccessKey)
	assert.Equal(t, "testsecret", dbCfg.SecretAccessKey)
}
