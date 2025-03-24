package meetuptoken

import (
	"encoding/base64"
	"errors"
	"github.com/spf13/viper"
	"log"
	"strings"
)

const (
	meetupPrivateKeyBase64Key = "MEETUP_PRIVATE_KEY_BASE64"
	meetupPrivateKeyKey       = "MEETUP_PRIVATE_KEY"
	meetupUserIdKey           = "MEETUP_USER_ID"
	meetupClientKeyKey        = "MEETUP_CLIENT_KEY"
	meetupSigningKeyIdKey     = "MEETUP_SIGNING_KEY_ID"
	meetupAuthUrlKey          = "MEETUP_AUTH_URL"
)

type Config struct {
	MeetupPrivateKey   []byte `mapstructure:"meetup_private_key"`
	MeetupUserId       string `mapstructure:"meetup_user_id"`
	MeetupClientKey    string `mapstructure:"meetup_client_key"`
	MeetupSigningKeyId string `mapstructure:"meetup_signing_key_id"`
	MeetupAuthUrl      string `mapstructure:"meetup_auth_url"`
}

func LoadConfigFromEnvFile(path, filename string) *Config {
	v := viper.New()

	v.SetDefault(strings.ToLower(meetupPrivateKeyBase64Key), "")
	v.SetDefault(strings.ToLower(meetupPrivateKeyKey), "")
	v.SetDefault(strings.ToLower(meetupUserIdKey), "")
	v.SetDefault(strings.ToLower(meetupClientKeyKey), "")
	v.SetDefault(strings.ToLower(meetupSigningKeyIdKey), "")
	v.SetDefault(strings.ToLower(meetupAuthUrlKey), "https://secure.meetup.com/oauth2/access")

	v.SetConfigName(filename)
	v.SetConfigType("env")
	v.AddConfigPath(path)

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			log.Printf("Warning: error reading .env file: %v", err)
		}
	}

	v.AutomaticEnv()

	meetupPrivateKeyBase64 := v.Get(strings.ToLower(meetupPrivateKeyBase64Key)).(string)
	meetupPrivateKey, err := base64.StdEncoding.DecodeString(meetupPrivateKeyBase64)
	if err == nil {
		v.SetDefault(strings.ToLower(meetupPrivateKeyKey), meetupPrivateKey)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Printf("Unable to decode into struct: %v", err)
	}

	validateConfig(&cfg)

	return &cfg
}

func LoadConfig() *Config {
	return LoadConfigFromEnvFile(".", ".env")
}

func validateConfig(cfg *Config) {
	var missing []string
	if len(cfg.MeetupPrivateKey) == 0 {
		missing = append(missing, meetupPrivateKeyBase64Key)
	}
	if cfg.MeetupUserId == "" {
		missing = append(missing, meetupUserIdKey)
	}
	if cfg.MeetupClientKey == "" {
		missing = append(missing, meetupClientKeyKey)
	}
	if cfg.MeetupSigningKeyId == "" {
		missing = append(missing, meetupSigningKeyIdKey)
	}
	if cfg.MeetupAuthUrl == "" {
		missing = append(missing, meetupAuthUrlKey)
	}

	if len(missing) > 0 {
		log.Fatalf("Missing required env vars: %v", strings.Join(missing, ", "))
	}
}
