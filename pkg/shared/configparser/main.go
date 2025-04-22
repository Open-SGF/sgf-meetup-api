package configparser

import (
	"errors"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

type ParseOptions struct {
	EnvFilename string
	EnvFilepath string
	Keys        []string
	SetDefaults func(v *viper.Viper) error
}

func Parse[T any](options ParseOptions) (*T, error) {
	v := viper.New()

	for _, key := range options.Keys {
		v.SetDefault(strings.ToLower(key), "")
	}

	v.SetConfigName(options.EnvFilename)
	v.SetConfigType("env")
	v.AddConfigPath(options.EnvFilepath)

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	v.AutomaticEnv()

	if err := options.SetDefaults(v); err != nil {
		return nil, err
	}

	var cfg T
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ParseFromKey[T any](v *viper.Viper, key string, parser func(string) (T, error), fallback T) {
	normalizedKey := strings.ToLower(key)
	str := v.GetString(normalizedKey)
	value, err := parser(str)

	if err != nil {
		v.Set(normalizedKey, fallback)
		return
	}

	v.Set(normalizedKey, value)
}

func SetupTestEnv(envContent string) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "configtest")

	if err != nil {
		return "", nil, err
	}

	envPath := filepath.Join(tempDir, ".env")
	err = os.WriteFile(envPath, []byte(envContent), 0644)

	if err != nil {
		return "", nil, err
	}

	return tempDir, func() {
		_ = os.RemoveAll(tempDir)
	}, nil
}
