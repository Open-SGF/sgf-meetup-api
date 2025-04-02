package configparser

import (
	"errors"
	"github.com/spf13/viper"
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
