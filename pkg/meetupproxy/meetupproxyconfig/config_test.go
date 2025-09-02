package meetupproxyconfig

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sgf-meetup-api/pkg/shared/appconfig"
)

func TestNewConfig(t *testing.T) {
	awsConfigManager := appconfig.NewAwsConfigManager()
	ctx := context.Background()

	t.Run("successful load from environment variables", func(t *testing.T) {
		switchToTempTestDir(t)
		validKey := base64.StdEncoding.EncodeToString([]byte("private_key"))
		t.Setenv(meetupPrivateKeyBase64Key, validKey)
		t.Setenv(meetupUserIdKey, "user123")
		t.Setenv(meetupClientKeyKey, "client123")
		t.Setenv(meetupSigningKeyIdKey, "signing123")

		cfg, err := NewConfig(ctx, awsConfigManager)
		require.NoError(t, err)

		assert.Equal(t, "user123", cfg.MeetupUserID)
		assert.Equal(t, []byte("private_key"), cfg.MeetupPrivateKey)
		assert.Equal(t, "https://secure.meetup.com/oauth2/access", cfg.MeetupAuthURL)
		assert.Equal(t, "https://api.meetup.com/gql", cfg.MeetupAPIURL)
	})

	t.Run("successful load from .env file", func(t *testing.T) {
		tempDir := t.TempDir()
		envPath := filepath.Join(tempDir, ".env")
		validKey := base64.StdEncoding.EncodeToString([]byte("env_file_key"))

		envContent := strings.Join([]string{
			meetupPrivateKeyBase64Key + "=" + validKey,
			meetupUserIdKey + "=env_user",
			meetupClientKeyKey + "=env_client",
			meetupSigningKeyIdKey + "=env_signing",
		}, "\n")

		require.NoError(t, os.WriteFile(envPath, []byte(envContent), 0o600))
		switchToTempTestDir(t, tempDir)

		cfg, err := NewConfig(ctx, awsConfigManager)
		require.NoError(t, err)

		assert.Equal(t, "env_user", cfg.MeetupUserID)
		assert.Equal(t, []byte("env_file_key"), cfg.MeetupPrivateKey)
	})

	t.Run("validation fails with missing required fields", func(t *testing.T) {
		switchToTempTestDir(t)
		t.Setenv(meetupPrivateKeyBase64Key, "valid_base64=")

		_, err := NewConfig(ctx, awsConfigManager)
		require.Error(t, err)
		assert.Contains(t, err.Error(), meetupUserIdKey)
		assert.Contains(t, err.Error(), meetupClientKeyKey)
	})

	t.Run("invalid base64 in private key", func(t *testing.T) {
		switchToTempTestDir(t)
		t.Setenv(meetupPrivateKeyBase64Key, "invalid_base64")
		t.Setenv(meetupUserIdKey, "user123")
		t.Setenv(meetupClientKeyKey, "client123")
		t.Setenv(meetupSigningKeyIdKey, "signing123")

		_, err := NewConfig(ctx, awsConfigManager)
		require.Error(t, err)
		assert.Contains(t, err.Error(), meetupPrivateKeyBase64Key)
	})
}

func switchToTempTestDir(t *testing.T, customDir ...string) {
	t.Helper()

	tempDir := t.TempDir()
	if len(customDir) > 0 {
		tempDir = customDir[0]
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	require.NoError(t, os.Chdir(tempDir))
}
