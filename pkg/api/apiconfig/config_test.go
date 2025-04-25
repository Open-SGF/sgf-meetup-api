package apiconfig

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sgf-meetup-api/pkg/shared/appconfig"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	awsConfigManager := appconfig.NewAwsConfigManager()
	ctx := context.Background()

	t.Run("successful load from environment variables", func(t *testing.T) {
		switchToTempTestDir(t)
		t.Setenv("EVENTS_TABLE_NAME", "test_events")
		t.Setenv("API_USERS_TABLE_NAME", "test_api_users")
		t.Setenv("GROUP_ID_DATE_TIME_INDEX_NAME", "test_index")
		t.Setenv("JWT_SECRET", "test_secret")
		t.Setenv("APP_URL", "https://test.example.com")

		cfg, err := NewConfig(ctx, awsConfigManager)
		require.NoError(t, err)

		assert.Equal(t, "test_events", cfg.EventsTableName)
		assert.Equal(t, "test_api_users", cfg.ApiUsersTableName)
		assert.Equal(t, "test_index", cfg.GroupIDDateTimeIndexName)
		assert.Equal(t, []byte("test_secret"), cfg.JWTSecret)
		assert.Equal(t, "https://test.example.com", cfg.AppURL.String())
	})

	t.Run("successful load from .env file", func(t *testing.T) {
		tempDir := t.TempDir()
		envPath := filepath.Join(tempDir, ".env")

		envContent := strings.Join([]string{
			"EVENTS_TABLE_NAME=file_events",
			"API_USERS_TABLE_NAME=file_api_users",
			"GROUP_ID_DATE_TIME_INDEX_NAME=file_index",
			"JWT_SECRET=file_secret",
			"APP_URL=https://file.example.com",
		}, "\n")

		require.NoError(t, os.WriteFile(envPath, []byte(envContent), 0600))

		origDir, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(origDir) })

		require.NoError(t, os.Chdir(tempDir))

		cfg, err := NewConfig(ctx, awsConfigManager)
		require.NoError(t, err)

		assert.Equal(t, "file_events", cfg.EventsTableName)
		assert.Equal(t, "file_api_users", cfg.ApiUsersTableName)
		assert.Equal(t, "file_index", cfg.GroupIDDateTimeIndexName)
		assert.Equal(t, []byte("file_secret"), cfg.JWTSecret)
		assert.Equal(t, "https://file.example.com", cfg.AppURL.String())
	})

	t.Run("sets default values", func(t *testing.T) {
		switchToTempTestDir(t)
		t.Setenv("EVENTS_TABLE_NAME", "default_events")
		t.Setenv("API_USERS_TABLE_NAME", "default_api_users")
		t.Setenv("GROUP_ID_DATE_TIME_INDEX_NAME", "default_index")
		t.Setenv("JWT_SECRET", "default_secret")

		cfg, err := NewConfig(ctx, awsConfigManager)
		require.NoError(t, err)

		assert.Equal(t, "meetup-api.opensgf.org", cfg.JWTIssuer)
		assert.Equal(t, "https://meetup-api.opensgf.org", cfg.AppURL.String())
	})

	t.Run("validation fails with missing fields", func(t *testing.T) {
		switchToTempTestDir(t)

		_, err := NewConfig(ctx, awsConfigManager)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "EVENTS_TABLE_NAME")
	})

	t.Run("invalid app URL format", func(t *testing.T) {
		switchToTempTestDir(t)
		t.Setenv("EVENTS_TABLE_NAME", "invalid_url_events")
		t.Setenv("API_USERS_TABLE_NAME", "invalid_url_api_users")
		t.Setenv("GROUP_ID_DATE_TIME_INDEX_NAME", "invalid_url_index")
		t.Setenv("JWT_SECRET", "invalid_url_secret")
		t.Setenv("APP_URL", "://invalid.url")

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
