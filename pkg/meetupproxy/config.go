package meetupproxy

import (
	"encoding/base64"
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"sgf-meetup-api/pkg/configparser"
	"strings"
)

const (
	logLevelKey               = "LOG_LEVEL"
	sentryDsnKey              = "SENTRY_DSN"
	meetupPrivateKeyBase64Key = "MEETUP_PRIVATE_KEY_BASE64"
	meetupPrivateKeyKey       = "MEETUP_PRIVATE_KEY"
	meetupUserIdKey           = "MEETUP_USER_ID"
	meetupClientKeyKey        = "MEETUP_CLIENT_KEY"
	meetupSigningKeyIdKey     = "MEETUP_SIGNING_KEY_ID"
	meetupAuthUrlKey          = "MEETUP_AUTH_URL"
	meetupApiUrlKey           = "MEETUP_API_URL"
)

var keys = []string{
	strings.ToLower(logLevelKey),
	strings.ToLower(sentryDsnKey),
	strings.ToLower(meetupPrivateKeyBase64Key),
	strings.ToLower(meetupPrivateKeyKey),
	strings.ToLower(meetupUserIdKey),
	strings.ToLower(meetupClientKeyKey),
	strings.ToLower(meetupSigningKeyIdKey),
	strings.ToLower(meetupAuthUrlKey),
	strings.ToLower(meetupApiUrlKey),
}

type Config struct {
	LogLevel           slog.Level `mapstructure:"log_level"`
	SentryDsn          string     `mapstructure:"sentry_dsn"`
	MeetupPrivateKey   []byte     `mapstructure:"meetup_private_key"`
	MeetupUserID       string     `mapstructure:"meetup_user_id"`
	MeetupClientKey    string     `mapstructure:"meetup_client_key"`
	MeetupSigningKeyID string     `mapstructure:"meetup_signing_key_id"`
	MeetupAuthURL      string     `mapstructure:"meetup_auth_url"`
	MeetupAPIURL       string     `mapstructure:"meetup_api_url"`
}

func NewConfig() (*Config, error) {
	return NewConfigFromEnvFile(".", ".env")
}

func NewConfigFromEnvFile(path, filename string) (*Config, error) {
	config, err := configparser.Parse[Config](configparser.ParseOptions{
		EnvFilepath: path,
		EnvFilename: filename,
		Keys:        keys,
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

func setDefaults(v *viper.Viper) {
	configparser.ParseLogLevelFromKey(v, strings.ToLower(logLevelKey), slog.LevelInfo)
	v.SetDefault(strings.ToLower(meetupAuthUrlKey), "https://secure.meetup.com/oauth2/access")
	v.SetDefault(strings.ToLower(meetupApiUrlKey), "https://api.meetup.com/gql")

	meetupPrivateKeyBase64 := v.Get(strings.ToLower(meetupPrivateKeyBase64Key)).(string)
	meetupPrivateKey, err := base64.StdEncoding.DecodeString(meetupPrivateKeyBase64)
	if err == nil {
		v.SetDefault(strings.ToLower(meetupPrivateKeyKey), meetupPrivateKey)
	}
}

func (cfg *Config) validate() error {
	var missing []string
	if len(cfg.MeetupPrivateKey) == 0 {
		missing = append(missing, meetupPrivateKeyBase64Key)
	}
	if cfg.MeetupUserID == "" {
		missing = append(missing, meetupUserIdKey)
	}
	if cfg.MeetupClientKey == "" {
		missing = append(missing, meetupClientKeyKey)
	}
	if cfg.MeetupSigningKeyID == "" {
		missing = append(missing, meetupSigningKeyIdKey)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %v", strings.Join(missing, ", "))
	}

	return nil
}

func getLogLevel(cfg *Config) slog.Level {
	return cfg.LogLevel
}
