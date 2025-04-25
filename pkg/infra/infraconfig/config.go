package infraconfig

import (
	"context"
	"github.com/spf13/viper"
	"log/slog"
	"sgf-meetup-api/pkg/shared/appconfig"
	"sgf-meetup-api/pkg/shared/logging"
)

const (
	logLevelKey      = "LOG_LEVEL"
	logTypeKey       = "LOG_TYPE"
	appEnvKey        = "APP_ENV"
	appDomainNameEnv = "APP_DOMAIN_NAME"
)

var configKeys = []string{
	logLevelKey,
	logTypeKey,
	appEnvKey,
	appDomainNameEnv,
}

type Config struct {
	LogLevel      slog.Level      `mapstructure:"log_level"`
	LogType       logging.LogType `mapstructure:"log_type"`
	AppEnv        string          `mapstructure:"app_env"`
	AppDomainName string          `mapstructure:"app_domain_name"`
}

func NewConfig(ctx context.Context) (*Config, error) {
	return NewConfigFromEnvFile(ctx, ".", ".env")
}

func NewConfigFromEnvFile(ctx context.Context, path, filename string) (*Config, error) {
	config, err := appconfig.Parse[Config](ctx, appconfig.ParseOptions{
		EnvFilepath: path,
		EnvFilename: filename,
		Keys:        configKeys,
		SetDefaults: setDefaults,
	})

	if err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults(v *viper.Viper) error {
	appconfig.ParseFromKey(v, logLevelKey, logging.ParseLogLevel, slog.LevelInfo)
	appconfig.ParseFromKey(v, logTypeKey, logging.ParseLogType, logging.LogTypeText)
	return nil
}

func NewLoggingConfig(config *Config) logging.Config {
	return logging.Config{
		Level: config.LogLevel,
		Type:  config.LogType,
	}
}
