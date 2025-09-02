package syncdynamodbconfig

import (
	"context"

	"github.com/google/wire"
	"sgf-meetup-api/pkg/shared/appconfig"
)

type Config struct {
	appconfig.Common `mapstructure:",squash"`
}

func NewConfig(
	ctx context.Context,
	awsConfigFactory *appconfig.AwsConfigManagerImpl,
) (*Config, error) {
	var config Config

	err := appconfig.NewParser().
		WithCommonConfig().
		WithEnvFile(".", ".env").
		WithEnvVars().
		WithCustomProcessor(awsConfigFactory.SetConfigFromViper).
		Parse(ctx, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

var ConfigProviders = wire.NewSet(
	appconfig.ConfigProviders,
	wire.FieldsOf(new(*Config), "Common"),
	NewConfig,
)
