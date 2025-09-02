package upsertuserconfig

import (
	"context"
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
		t.Setenv(appconfig.DynamoDBEndpointKey, "dynamodb_endpoint")

		cfg, err := NewConfig(ctx, awsConfigManager)
		require.NoError(t, err)

		assert.Equal(t, "dynamodb_endpoint", cfg.DynamoDB.Endpoint)
	})

	t.Run("successful load from .env file", func(t *testing.T) {
		tempDir := t.TempDir()
		envPath := filepath.Join(tempDir, ".env")

		envContent := strings.Join([]string{
			appconfig.DynamoDBEndpointKey + "=dynamodb_endpoint",
		}, "\n")

		require.NoError(t, os.WriteFile(envPath, []byte(envContent), 0o600))

		origDir, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(origDir) })
		require.NoError(t, os.Chdir(tempDir))

		cfg, err := NewConfig(ctx, awsConfigManager)
		require.NoError(t, err)

		assert.Equal(t, "dynamodb_endpoint", cfg.DynamoDB.Endpoint)
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
