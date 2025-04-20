package configparser

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"sgf-meetup-api/pkg/shared/logging"
	"testing"
)

func TestParse_ValidEnvFile(t *testing.T) {
	tempDir, cleanup, err := SetupTestEnv(`
DB_HOST=localhost
DB_PORT=5432
`)

	require.NoError(t, err)
	defer cleanup()

	type Config struct {
		Host string `mapstructure:"db_host"`
		Port int    `mapstructure:"db_port"`
	}

	opts := ParseOptions{
		EnvFilename: ".env",
		EnvFilepath: tempDir,
		SetDefaults: func(v *viper.Viper) { v.SetDefault("db_port", 3306) },
	}

	cfg, err := Parse[Config](opts)

	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
}

func TestParse_EnvVarOverridesConfigFile(t *testing.T) {
	tempDir, cleanup, err := SetupTestEnv(`
API_KEY=file_key
`)

	require.NoError(t, err)
	defer cleanup()

	t.Setenv("API_KEY", "env_key")

	type Config struct {
		APIKey string `mapstructure:"api_key"`
	}

	opts := ParseOptions{
		EnvFilename: ".env",
		EnvFilepath: tempDir,
		SetDefaults: func(v *viper.Viper) {},
	}

	cfg, err := Parse[Config](opts)

	require.NoError(t, err)

	assert.Equal(t, "env_key", cfg.APIKey)
}

func TestParse_MissingConfigFileWithDefaults(t *testing.T) {
	type Config struct {
		LogLevel string `mapstructure:"log_level"`
	}

	opts := ParseOptions{
		EnvFilename: "missing.env",
		EnvFilepath: t.TempDir(),
		SetDefaults: func(v *viper.Viper) {
			v.SetDefault("log_level", "info")
		},
	}

	cfg, err := Parse[Config](opts)

	require.NoError(t, err)

	assert.Equal(t, "info", cfg.LogLevel)
}

func TestParse_InvalidConfigFile(t *testing.T) {
	tempDir, cleanup, err := SetupTestEnv(`
INVALID_KEY_WITHOUT_VALUE
`)

	require.NoError(t, err)
	defer cleanup()

	type Config struct {
		Key string `mapstructure:"INVALID_KEY_WITHOUT_VALUE"`
	}

	opts := ParseOptions{
		EnvFilename: ".env",
		EnvFilepath: tempDir,
		SetDefaults: func(v *viper.Viper) {},
	}

	_, err = Parse[Config](opts)

	require.Error(t, err)
}

func TestParse_EmptyKeysInitialization(t *testing.T) {
	type Config struct {
		FeatureFlag string `mapstructure:"feature_flag"`
	}

	opts := ParseOptions{
		Keys:        []string{"feature_flag"},
		SetDefaults: func(v *viper.Viper) {},
	}

	cfg, err := Parse[Config](opts)

	require.NoError(t, err)

	assert.Equal(t, "", cfg.FeatureFlag)
}

func TestParseFromKey_LogLevel(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		fallbackLvl slog.Level
		expectedLvl slog.Level
	}{
		{"correct string value", "DEBUG", slog.LevelInfo, slog.LevelDebug},
		{"incorrect string value", "invalid", slog.LevelWarn, slog.LevelWarn},
		{"non-string value", 0, slog.LevelError, slog.LevelError},
		{"nil value", nil, slog.LevelInfo, slog.LevelInfo},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key := "loglevel"
			v := viper.New()
			v.SetDefault(key, tc.input)

			ParseFromKey(v, key, logging.ParseLogLevel, tc.fallbackLvl)

			value := v.Get(key)

			assert.IsType(t, slog.LevelInfo, value)
			assert.Equal(t, value, tc.expectedLvl)
		})
	}
}
