package importerconfig

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
SENTRY_DSN=https://sentry.example.com
MEETUP_GROUP_NAMES=test1,test2
DYNAMODB_ENDPOINT=http://localhost:8000
AWS_REGION=us-west-2
AWS_ACCESS_KEY=testkey
AWS_SECRET_ACCESS_KEY=testsecret
MEETUP_PROXY_FUNCTION_NAME=meetupproxy
ARCHIVED_EVENTS_TABLE_NAME=archived-events
EVENTS_TABLE_NAME=events
GROUP_ID_DATE_TIME_INDEX_NAME=group-index
`)

		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(context.Background(), tempDir, ".env")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, slog.LevelDebug, cfg.LogLevel)
		assert.Equal(t, logging.LogTypeJSON, cfg.LogType)
		assert.Equal(t, "https://sentry.example.com", cfg.SentryDsn)
		assert.ElementsMatch(t, []string{"test1", "test2"}, cfg.MeetupGroupNames)
		assert.Equal(t, "http://localhost:8000", cfg.DynamoDbEndpoint)
		assert.Equal(t, "us-west-2", cfg.AwsRegion)
		assert.Equal(t, "testkey", cfg.AwsAccessKey)
		assert.Equal(t, "testsecret", cfg.AwsSecretAccessKey)
		assert.Equal(t, "meetupproxy", cfg.ProxyFunctionName)
		assert.Equal(t, "archived-events", cfg.ArchivedEventsTableName)
		assert.Equal(t, "events", cfg.EventsTableName)
		assert.Equal(t, "group-index", cfg.GroupIDDateTimeIndexName)
	})

	t.Run("minimal with defaults", func(t *testing.T) {
		tempDir, cleanup, err := appconfig.SetupTestEnv(`
MEETUP_PROXY_FUNCTION_NAME=meetupproxy
ARCHIVED_EVENTS_TABLE_NAME=archived-events
EVENTS_TABLE_NAME=events
GROUP_ID_DATE_TIME_INDEX_NAME=group-index
`)

		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(context.Background(), tempDir, ".env")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, slog.LevelInfo, cfg.LogLevel)
		assert.Equal(t, logging.LogTypeText, cfg.LogType)
		assert.Equal(t, "us-east-2", cfg.AwsRegion)
		assert.Equal(t, []string{}, cfg.MeetupGroupNames)
	})

	t.Run("invalid missing required", func(t *testing.T) {
		tempDir, cleanup, err := appconfig.SetupTestEnv(`
LOG_LEVEL=info
`)
		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(context.Background(), tempDir, ".env")
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), proxyFunctionNameKey)
		assert.Contains(t, err.Error(), archivedEventsTableNameKey)
		assert.Contains(t, err.Error(), eventsTableNameKey)
		assert.Contains(t, err.Error(), groupIDDateTimeIndexNameKey)
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
