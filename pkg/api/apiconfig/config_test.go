package apiconfig

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sgf-meetup-api/pkg/shared/appconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	awsConfigManager := appconfig.NewAwsConfigManager()
	ctx := context.Background()

	t.Run("successful load from environment variables", func(t *testing.T) {
		switchToTempTestDir(t)
		t.Setenv(eventsTableNameKey, "test_events")
		t.Setenv(apiUsersTableNameKey, "test_api_users")
		t.Setenv(groupIDDateTimeIndexNameKey, "test_index")
		t.Setenv(jwtSecretBase64Key, "dGVzdF9iYXNlNjQ=")
		t.Setenv(appUrlKey, "https://test.example.com")

		cfg, err := NewConfig(ctx, awsConfigManager)
		require.NoError(t, err)

		assert.Equal(t, "test_events", cfg.EventsTableName)
		assert.Equal(t, "test_api_users", cfg.APIUsersTableName)
		assert.Equal(t, "test_index", cfg.GroupIDDateTimeIndexName)
		assert.Equal(t, []byte("test_base64"), cfg.JWTSecret)
		assert.Equal(t, "https://test.example.com", cfg.AppURL.String())
	})

	t.Run("successful load from .env file", func(t *testing.T) {
		tempDir := t.TempDir()
		envPath := filepath.Join(tempDir, ".env")

		envContent := strings.Join([]string{
			eventsTableNameKey + "=file_events",
			apiUsersTableNameKey + "=file_api_users",
			groupIDDateTimeIndexNameKey + "=file_index",
			jwtSecretBase64Key + "=dGVzdF9iYXNlNjQ=",
			appUrlKey + "=https://file.example.com",
		}, "\n")

		require.NoError(t, os.WriteFile(envPath, []byte(envContent), 0o600))

		origDir, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(origDir) })

		require.NoError(t, os.Chdir(tempDir))

		cfg, err := NewConfig(ctx, awsConfigManager)
		require.NoError(t, err)

		assert.Equal(t, "file_events", cfg.EventsTableName)
		assert.Equal(t, "file_api_users", cfg.APIUsersTableName)
		assert.Equal(t, "file_index", cfg.GroupIDDateTimeIndexName)
		assert.Equal(t, []byte("test_base64"), cfg.JWTSecret)
		assert.Equal(t, "https://file.example.com", cfg.AppURL.String())
	})

	t.Run("sets default values", func(t *testing.T) {
		switchToTempTestDir(t)
		t.Setenv(eventsTableNameKey, "default_events")
		t.Setenv(apiUsersTableNameKey, "default_api_users")
		t.Setenv(groupIDDateTimeIndexNameKey, "default_index")
		t.Setenv(jwtSecretKey, "default_secret")

		cfg, err := NewConfig(ctx, awsConfigManager)
		require.NoError(t, err)

		assert.Equal(t, "sgf-meetup-api.opensgf.org", cfg.JWTIssuer)
		assert.Equal(t, "https://sgf-meetup-api.opensgf.org", cfg.AppURL.String())
	})

	t.Run("validation fails with missing fields", func(t *testing.T) {
		switchToTempTestDir(t)

		_, err := NewConfig(ctx, awsConfigManager)
		require.Error(t, err)
		assert.Contains(t, err.Error(), eventsTableNameKey)
	})

	t.Run("invalid app URL format", func(t *testing.T) {
		switchToTempTestDir(t)
		t.Setenv(eventsTableNameKey, "invalid_url_events")
		t.Setenv(apiUsersTableNameKey, "invalid_url_api_users")
		t.Setenv(groupIDDateTimeIndexNameKey, "invalid_url_index")
		t.Setenv(jwtSecretKey, "invalid_url_secret")
		t.Setenv(appUrlKey, "://invalid.url")

		_, err := NewConfig(ctx, awsConfigManager)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse")
	})
}

func switchToTempTestDir(t *testing.T) {
	t.Helper()

	originalDir, err := os.Getwd()
	require.NoError(t, err)

	tempDir := t.TempDir()
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	require.NoError(t, os.Chdir(tempDir))
}
