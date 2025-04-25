package meetupproxyconfig

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
MEETUP_PRIVATE_KEY_BASE64=c29tZUJhc2U2NEtleQ==
MEETUP_USER_ID=meetupUserId
MEETUP_CLIENT_KEY=meetupClientKey
MEETUP_SIGNING_KEY_ID=signingKeyId
MEETUP_AUTH_URL=https://api.example.com/auth
MEETUP_API_URL=https://api.example.com
`)

		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(context.Background(), tempDir, ".env")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, slog.LevelDebug, cfg.LogLevel)
		assert.Equal(t, logging.LogTypeJSON, cfg.LogType)
		assert.Equal(t, "https://sentry.example.com", cfg.SentryDsn)
		assert.Equal(t, []byte("someBase64Key"), cfg.MeetupPrivateKey)
		assert.Equal(t, "meetupUserId", cfg.MeetupUserID)
		assert.Equal(t, "meetupClientKey", cfg.MeetupClientKey)
		assert.Equal(t, "signingKeyId", cfg.MeetupSigningKeyID)
		assert.Equal(t, "https://api.example.com/auth", cfg.MeetupAuthURL)
		assert.Equal(t, "https://api.example.com", cfg.MeetupAPIURL)
	})

	t.Run("minimal with defaults", func(t *testing.T) {
		tempDir, cleanup, err := appconfig.SetupTestEnv(`
MEETUP_PRIVATE_KEY_BASE64=c29tZUJhc2U2NEtleQ==
MEETUP_USER_ID=meetupUserId
MEETUP_CLIENT_KEY=meetupClientKey
MEETUP_SIGNING_KEY_ID=signingKeyId
`)

		require.NoError(t, err)
		defer cleanup()

		cfg, err := NewConfigFromEnvFile(context.Background(), tempDir, ".env")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, slog.LevelInfo, cfg.LogLevel)
		assert.Equal(t, logging.LogTypeText, cfg.LogType)
		assert.Equal(t, "https://secure.meetup.com/oauth2/access", cfg.MeetupAuthURL)
		assert.Equal(t, "https://api.meetup.com/gql", cfg.MeetupAPIURL)
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
		assert.Contains(t, err.Error(), meetupPrivateKeyBase64Key)
		assert.Contains(t, err.Error(), meetupUserIdKey)
		assert.Contains(t, err.Error(), meetupClientKeyKey)
		assert.Contains(t, err.Error(), meetupSigningKeyIdKey)
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
