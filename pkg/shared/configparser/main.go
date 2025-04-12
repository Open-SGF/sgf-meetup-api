package configparser

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"strings"
)

type ParseOptions struct {
	EnvFilename string
	EnvFilepath string
	Keys        []string
	SetDefaults func(v *viper.Viper)
}

func Parse[T any](options ParseOptions) (*T, error) {
	v := viper.New()

	for _, key := range options.Keys {
		v.SetDefault(key, "")
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

	options.SetDefaults(v)

	var cfg T
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ParseLogLevelFromKey(v *viper.Viper, key string, fallback slog.Level) {
	levelStr := v.GetString(key)
	level, err := ParseLogLevel(levelStr)

	if err != nil {
		v.Set(key, fallback)
		return
	}

	v.Set(key, level)
}

func ParseLogLevel(s string) (slog.Level, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN", "WARNING":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unknown log level: %q", s)
	}
}
