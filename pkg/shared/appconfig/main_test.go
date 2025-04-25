package appconfig

import (
	"context"
	"errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"sgf-meetup-api/pkg/shared/logging"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("valid env file", func(t *testing.T) {
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
			SetDefaults: func(v *viper.Viper) error {
				v.SetDefault("db_port", 3306)
				return nil
			},
		}

		cfg, err := Parse[Config](context.Background(), opts)

		require.NoError(t, err)

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
	})

	t.Run("env var overrides .env file", func(t *testing.T) {
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
			SetDefaults: func(v *viper.Viper) error { return nil },
		}

		cfg, err := Parse[Config](context.Background(), opts)

		require.NoError(t, err)

		assert.Equal(t, "env_key", cfg.APIKey)
	})

	t.Run("missing config file with defaults", func(t *testing.T) {
		type Config struct {
			LogLevel string `mapstructure:"log_level"`
		}

		opts := ParseOptions{
			EnvFilename: "missing.env",
			EnvFilepath: t.TempDir(),
			SetDefaults: func(v *viper.Viper) error {
				v.SetDefault("log_level", "info")
				return nil
			},
		}

		cfg, err := Parse[Config](context.Background(), opts)

		require.NoError(t, err)

		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("invalid config file", func(t *testing.T) {
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
			SetDefaults: func(v *viper.Viper) error { return nil },
		}

		_, err = Parse[Config](context.Background(), opts)

		require.Error(t, err)
	})

	t.Run("empty keys initialization", func(t *testing.T) {
		type Config struct {
			FeatureFlag string `mapstructure:"feature_flag"`
		}

		opts := ParseOptions{
			Keys:        []string{"feature_flag"},
			SetDefaults: func(v *viper.Viper) error { return nil },
		}

		cfg, err := Parse[Config](context.Background(), opts)

		require.NoError(t, err)

		assert.Equal(t, "", cfg.FeatureFlag)
	})

	t.Run("handle SetDefaults error", func(t *testing.T) {
		type Config struct{}

		errSetDefaults := errors.New("set defaults error")

		opts := ParseOptions{
			Keys: []string{},
			SetDefaults: func(v *viper.Viper) error {
				return errSetDefaults
			},
		}

		_, err := Parse[Config](context.Background(), opts)

		assert.ErrorIs(t, err, errSetDefaults)
	})
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
