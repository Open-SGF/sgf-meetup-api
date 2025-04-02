package configparser

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
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
