package upsertuserconfig

import (
	"context"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/shared/appconfig"
)

const (
	appEnvKey = "APP_ENV"
)

var configKeys = []string{
	appEnvKey,
}

type Config struct {
	appconfig.Common `mapstructure:",squash"`
	AppEnv           string `mapstructure:"app_env"`
}

func NewConfig(ctx context.Context, awsConfigFactory appconfig.AwsConfigManager) (*Config, error) {
	var config Config

	err := appconfig.NewParser().
		WithCommonConfig().
		DefineKeys(configKeys).
		WithEnvFile(".", ".env").
		WithEnvVars().
		WithCustomProcessor(awsConfigFactory.SetConfigFromViper).
		Parse(ctx, &config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

var ConfigProviders = wire.NewSet(appconfig.ConfigProviders, wire.FieldsOf(new(*Config), "Common"), NewConfig)
