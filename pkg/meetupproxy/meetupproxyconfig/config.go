package meetupproxyconfig

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"sgf-meetup-api/pkg/shared/appconfig"
	"strings"
)

const (
	meetupPrivateKeyBase64Key = "MEETUP_PRIVATE_KEY_BASE64"
	meetupPrivateKeyKey       = "MEETUP_PRIVATE_KEY"
	meetupUserIdKey           = "MEETUP_USER_ID"
	meetupClientKeyKey        = "MEETUP_CLIENT_KEY"
	meetupSigningKeyIdKey     = "MEETUP_SIGNING_KEY_ID"
	meetupAuthUrlKey          = "MEETUP_AUTH_URL"
	meetupApiUrlKey           = "MEETUP_API_URL"
)

var configKeys = []string{
	meetupPrivateKeyBase64Key,
	meetupPrivateKeyKey,
	meetupUserIdKey,
	meetupClientKeyKey,
	meetupSigningKeyIdKey,
	meetupAuthUrlKey,
	meetupApiUrlKey,
}

type Config struct {
	appconfig.Common   `mapstructure:",squash"`
	MeetupPrivateKey   []byte `mapstructure:"meetup_private_key"`
	MeetupUserID       string `mapstructure:"meetup_user_id"`
	MeetupClientKey    string `mapstructure:"meetup_client_key"`
	MeetupSigningKeyID string `mapstructure:"meetup_signing_key_id"`
	MeetupAuthURL      string `mapstructure:"meetup_auth_url"`
	MeetupAPIURL       string `mapstructure:"meetup_api_url"`
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
	v.SetDefault(strings.ToLower(meetupAuthUrlKey), "https://secure.meetup.com/oauth2/access")
	v.SetDefault(strings.ToLower(meetupApiUrlKey), "https://api.meetup.com/gql")

	meetupPrivateKeyBase64 := v.Get(strings.ToLower(meetupPrivateKeyBase64Key)).(string)
	meetupPrivateKey, err := base64.StdEncoding.DecodeString(meetupPrivateKeyBase64)
	if err == nil {
		v.SetDefault(strings.ToLower(meetupPrivateKeyKey), meetupPrivateKey)
	}
	return nil
}

func (config *Config) validate() error {
	var missing []string
	if len(config.MeetupPrivateKey) == 0 {
		missing = append(missing, meetupPrivateKeyBase64Key)
	}
	if config.MeetupUserID == "" {
		missing = append(missing, meetupUserIdKey)
	}
	if config.MeetupClientKey == "" {
		missing = append(missing, meetupClientKeyKey)
	}
	if config.MeetupSigningKeyID == "" {
		missing = append(missing, meetupSigningKeyIdKey)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %v", strings.Join(missing, ", "))
	}

	return nil
}

var ConfigProviders = wire.NewSet(appconfig.ConfigProviders, wire.FieldsOf(new(*Config), "Common"), NewConfig)
