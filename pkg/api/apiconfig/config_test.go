package apiconfig

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"sgf-meetup-api/pkg/shared/configparser"
	"sgf-meetup-api/pkg/shared/logging"
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Run("all values", func(t *testing.T) {
		tempDir, cleanup, err := configparser.SetupTestEnv(`
LOG_LEVEL=debug
LOG_TYPE=json
SENTRY_DSN=https://sentry.example.com
DYNAMODB_ENDPOINT=http://localhost:8000
AWS_REGION=us-west-2
AWS_ACCESS_KEY=testkey
AWS_SECRET_ACCESS_KEY=testsecret
EVENTS_TABLE_NAME=events
API_USERS_TABLE_NAME=users
GROUP_ID_DATE_TIME_INDEX_NAME=group-index
JWT_ISSUER=myapp
JWT_SECRET=secretkey
APP_URL=http://localhost
`)

		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(tempDir, ".env")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, slog.LevelDebug, cfg.LogLevel)
		assert.Equal(t, logging.LogTypeJSON, cfg.LogType)
		assert.Equal(t, "https://sentry.example.com", cfg.SentryDsn)
		assert.Equal(t, "http://localhost:8000", cfg.DynamoDbEndpoint)
		assert.Equal(t, "us-west-2", cfg.AwsRegion)
		assert.Equal(t, "testkey", cfg.AwsAccessKey)
		assert.Equal(t, "testsecret", cfg.AwsSecretAccessKey)
		assert.Equal(t, "events", cfg.EventsTableName)
		assert.Equal(t, "users", cfg.ApiUsersTableName)
		assert.Equal(t, "group-index", cfg.GroupIDDateTimeIndexName)
		assert.Equal(t, "myapp", cfg.JWTIssuer)
		assert.Equal(t, []byte("secretkey"), cfg.JWTSecret)
		assert.Equal(t, "http://localhost", cfg.AppURL.String())
	})

	t.Run("minimal with defaults", func(t *testing.T) {
		tempDir, cleanup, err := configparser.SetupTestEnv(`
EVENTS_TABLE_NAME=events
API_USERS_TABLE_NAME=users
GROUP_ID_DATE_TIME_INDEX_NAME=group-index
JWT_SECRET=secretkey
`)

		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(tempDir, ".env")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, slog.LevelInfo, cfg.LogLevel)
		assert.Equal(t, logging.LogTypeText, cfg.LogType)
		assert.Equal(t, "us-east-2", cfg.AwsRegion)
		assert.Equal(t, "meetup-api.opensgf.org", cfg.JWTIssuer)
		assert.Equal(t, "https://meetup-api.opensgf.org", cfg.AppURL.String())
	})

	t.Run("invalid missing required", func(t *testing.T) {
		tempDir, cleanup, err := configparser.SetupTestEnv(`
LOG_LEVEL=info
`)
		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(tempDir, ".env")
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), apiUsersTableNameKey)
		assert.Contains(t, err.Error(), eventsTableNameKey)
		assert.Contains(t, err.Error(), groupIDDateTimeIndexNameKey)
		assert.Contains(t, err.Error(), jwtSecretKey)
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
