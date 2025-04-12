package configparser

import (
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse_ValidEnvFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("DB_HOST=localhost\nDB_PORT=5432"), 0644); err != nil {
		t.Fatal(err)
	}

	type Config struct {
		Host string `mapstructure:"db_host"`
		Port int    `mapstructure:"db_port"`
	}

	opts := ParseOptions{
		EnvFilename: ".env",
		EnvFilepath: dir,
		SetDefaults: func(v *viper.Viper) { v.SetDefault("db_port", 3306) },
	}

	cfg, err := Parse[Config](opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Expected Host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("Expected Port 5432, got %d", cfg.Port)
	}
}

func TestParse_EnvVarOverridesConfigFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("API_KEY=file_key"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("API_KEY", "env_key")

	type Config struct {
		APIKey string `mapstructure:"API_KEY"`
	}

	opts := ParseOptions{
		EnvFilename: ".env",
		EnvFilepath: dir,
		SetDefaults: func(v *viper.Viper) {},
	}

	cfg, err := Parse[Config](opts)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.APIKey != "env_key" {
		t.Errorf("Expected API_KEY 'env_key', got '%s'", cfg.APIKey)
	}
}

func TestParse_MissingConfigFileWithDefaults(t *testing.T) {
	type Config struct {
		LogLevel string `mapstructure:"LOG_LEVEL"`
	}

	opts := ParseOptions{
		EnvFilename: "missing.env",
		EnvFilepath: t.TempDir(),
		SetDefaults: func(v *viper.Viper) {
			v.SetDefault("LOG_LEVEL", "info")
		},
	}

	cfg, err := Parse[Config](opts)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel 'info', got '%s'", cfg.LogLevel)
	}
}

func TestParse_InvalidConfigFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("INVALID_KEY_WITHOUT_VALUE"), 0644); err != nil {
		t.Fatal(err)
	}

	type Config struct {
		Key string `mapstructure:"INVALID_KEY_WITHOUT_VALUE"`
	}

	opts := ParseOptions{
		EnvFilename: ".env",
		EnvFilepath: dir,
		SetDefaults: func(v *viper.Viper) {},
	}

	_, err := Parse[Config](opts)
	if err == nil {
		t.Error("Expected error for invalid config file, got nil")
	}
}

func TestParse_EmptyKeysInitialization(t *testing.T) {
	type Config struct {
		FeatureFlag string `mapstructure:"FEATURE_FLAG"`
	}

	opts := ParseOptions{
		Keys:        []string{"FEATURE_FLAG"},
		SetDefaults: func(v *viper.Viper) {},
	}

	cfg, err := Parse[Config](opts)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.FeatureFlag != "" {
		t.Errorf("Expected empty FeatureFlag, got '%s'", cfg.FeatureFlag)
	}
}

func TestParseLogLevelFromKey(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		fallbackLvl slog.Level
		expectedLvl slog.Level
	}{
		{"correct string value", "DEBUG", slog.LevelInfo, slog.LevelDebug},
		{"incorrect string value", "some string", slog.LevelWarn, slog.LevelWarn},
		{"int value", 0, slog.LevelWarn, slog.LevelWarn},
		{"nil value", nil, slog.LevelWarn, slog.LevelWarn},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key := "key"
			v := viper.New()
			v.SetDefault(key, tc.input)

			ParseLogLevelFromKey(v, key, tc.fallbackLvl)

			level := v.Get(key).(slog.Level)

			if level != tc.expectedLvl {
				t.Fatalf("Expected level %v, got %v", tc.expectedLvl, level)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLvl slog.Level
		expectErr   bool
	}{
		{"debug lowercase", "debug", slog.LevelDebug, false},
		{"DEBUG uppercase", "DEBUG", slog.LevelDebug, false},
		{"info mixed case", "iNfO", slog.LevelInfo, false},
		{"warn alias", "warn", slog.LevelWarn, false},
		{"warning full", "WARNING", slog.LevelWarn, false},
		{"error with spaces", "  error  ", slog.LevelError, false},
		{"invalid level", "critical", 0, true},
		{"empty string", "", 0, true},
		{"numeric input", "123", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lvl, err := ParseLogLevel(tc.input)

			if tc.expectErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tc.input) {
					t.Errorf("Error message %q should contain input %q", err.Error(), tc.input)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if lvl != tc.expectedLvl {
					t.Errorf("Expected level %v, got %v", tc.expectedLvl, lvl)
				}
			}
		})
	}
}
