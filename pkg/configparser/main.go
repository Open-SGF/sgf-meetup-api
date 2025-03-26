package configparser

import (
	"errors"
	"github.com/spf13/viper"
	"log"
)

type ParseOptions struct {
	EnvFilename string
	EnvFilepath string
	Key         []string
	SetDefaults func(v *viper.Viper)
}

func Parse[T any](options ParseOptions) (*T, error) {
	v := viper.New()

	for _, key := range options.Key {
		v.SetDefault(key, "")
	}

	v.SetConfigName(options.EnvFilename)
	v.SetConfigType("env")
	v.AddConfigPath(options.EnvFilename)

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
		log.Printf("Unable to decode into struct: %v", err)
		return nil, err
	}

	return &cfg, nil
}
