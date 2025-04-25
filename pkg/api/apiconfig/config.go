package apiconfig

import (
	"context"
	"fmt"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"net/url"
	"sgf-meetup-api/pkg/shared/appconfig"
	"strings"
)

const (
	eventsTableNameKey          = "EVENTS_TABLE_NAME"
	apiUsersTableNameKey        = "API_USERS_TABLE_NAME"
	groupIDDateTimeIndexNameKey = "GROUP_ID_DATE_TIME_INDEX_NAME"
	jwtIssuerKey                = "JWT_ISSUER"
	jwtSecretKey                = "JWT_SECRET"
	appUrlKey                   = "APP_URL"
)

var configKeys = []string{
	eventsTableNameKey,
	apiUsersTableNameKey,
	groupIDDateTimeIndexNameKey,
	jwtIssuerKey,
	jwtSecretKey,
	appUrlKey,
}

type Config struct {
	appconfig.Common         `mapstructure:",squash"`
	EventsTableName          string  `mapstructure:"events_table_name"`
	ApiUsersTableName        string  `mapstructure:"api_users_table_name"`
	GroupIDDateTimeIndexName string  `mapstructure:"group_id_date_time_index_name"`
	JWTIssuer                string  `mapstructure:"jwt_issuer"`
	JWTSecret                []byte  `mapstructure:"jwt_secret"`
	AppURL                   url.URL `mapstructure:"app_url"`
}

func NewConfig(ctx context.Context, awsConfigFactory *appconfig.AwsConfigManager) (*Config, error) {
	var config Config

	err := appconfig.NewParser().
		WithCommonConfig().
		DefineKeys(configKeys).
		WithEnvFile(".", ".env").
		WithEnvVars().
		WithCustomProcessor(awsConfigFactory.SetConfigFromViper).
		WithSSMParameters(func(ctx context.Context, v *viper.Viper, opts *appconfig.SSMParameterOptions) {
			opts.AwsConfig = awsConfigFactory.Config()
			opts.SSMPath = v.GetString(appconfig.SSMPathKey)
		}).
		WithCustomProcessor(setDefaults).
		Parse(ctx, &config)

	if err != nil {
		return nil, err
	}

	if err = config.validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults(_ context.Context, v *viper.Viper) error {
	v.SetDefault(strings.ToLower(jwtIssuerKey), "meetup-api.opensgf.org")
	v.Set(strings.ToLower(jwtSecretKey), []byte(v.GetString(strings.ToLower(jwtSecretKey))))
	appUrl := v.GetString(strings.ToLower(appUrlKey))
	if appUrl == "" {
		appUrl = "https://meetup-api.opensgf.org"
	}
	parsedUrl, err := url.Parse(appUrl)

	if err != nil {
		return err
	}
	v.Set(strings.ToLower(appUrlKey), parsedUrl)

	return nil
}

func (config *Config) validate() error {
	var missing []string

	if config.EventsTableName == "" {
		missing = append(missing, eventsTableNameKey)
	}
	if config.ApiUsersTableName == "" {
		missing = append(missing, apiUsersTableNameKey)
	}
	if config.GroupIDDateTimeIndexName == "" {
		missing = append(missing, groupIDDateTimeIndexNameKey)
	}
	if len(config.JWTSecret) == 0 {
		missing = append(missing, jwtSecretKey)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %v", strings.Join(missing, ", "))
	}

	return nil
}

var ConfigProviders = wire.NewSet(appconfig.ConfigProviders, wire.FieldsOf(new(*Config), "Common"), NewConfig)
