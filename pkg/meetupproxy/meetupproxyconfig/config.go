package meetupproxyconfig

import (
	"encoding/base64"
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"sgf-meetup-api/pkg/shared/configparser"
	"sgf-meetup-api/pkg/shared/logging"
	"strings"
)

const (
	logLevelKey               = "LOG_LEVEL"
	logTypeKey                = "LOG_TYPE"
	sentryDsnKey              = "SENTRY_DSN"
	meetupPrivateKeyBase64Key = "MEETUP_PRIVATE_KEY_BASE64"
	meetupPrivateKeyKey       = "MEETUP_PRIVATE_KEY"
	meetupUserIdKey           = "MEETUP_USER_ID"
	meetupClientKeyKey        = "MEETUP_CLIENT_KEY"
	meetupSigningKeyIdKey     = "MEETUP_SIGNING_KEY_ID"
	meetupAuthUrlKey          = "MEETUP_AUTH_URL"
	meetupApiUrlKey           = "MEETUP_API_URL"
)

var configKeys = []string{
	logLevelKey,
	logTypeKey,
	sentryDsnKey,
	meetupPrivateKeyBase64Key,
	meetupPrivateKeyKey,
	meetupUserIdKey,
	meetupClientKeyKey,
	meetupSigningKeyIdKey,
	meetupAuthUrlKey,
	meetupApiUrlKey,
}

type Config struct {
	LogLevel           slog.Level      `mapstructure:"log_level"`
	LogType            logging.LogType `mapstructure:"log_type"`
	SentryDsn          string          `mapstructure:"sentry_dsn"`
	MeetupPrivateKey   []byte          `mapstructure:"meetup_private_key"`
	MeetupUserID       string          `mapstructure:"meetup_user_id"`
	MeetupClientKey    string          `mapstructure:"meetup_client_key"`
	MeetupSigningKeyID string          `mapstructure:"meetup_signing_key_id"`
	MeetupAuthURL      string          `mapstructure:"meetup_auth_url"`
	MeetupAPIURL       string          `mapstructure:"meetup_api_url"`
}

func NewConfig() (*Config, error) {
	return NewConfigFromEnvFile(".", ".env")
}

func NewConfigFromEnvFile(path, filename string) (*Config, error) {
	config, err := configparser.Parse[Config](configparser.ParseOptions{
		EnvFilepath: path,
		EnvFilename: filename,
		Keys:        configKeys,
		SetDefaults: setDefaults,
	})

	if err != nil {
		return nil, err
	}

	if err = config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults(v *viper.Viper) error {
	configparser.ParseFromKey(v, logLevelKey, logging.ParseLogLevel, slog.LevelInfo)
	configparser.ParseFromKey(v, logTypeKey, logging.ParseLogType, logging.LogTypeText)
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

func NewLoggingConfig(config *Config) logging.Config {
	return logging.Config{
		Level: config.LogLevel,
		Type:  config.LogType,
	}
}
