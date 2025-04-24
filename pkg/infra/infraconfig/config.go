package infraconfig

import (
	"github.com/spf13/viper"
	"log/slog"
	"sgf-meetup-api/pkg/shared/configparser"
	"sgf-meetup-api/pkg/shared/logging"
)

const (
	logLevelKey = "LOG_LEVEL"
	logTypeKey  = "LOG_TYPE"
	appEnvKey   = "APP_ENV"
)

var configKeys = []string{
	logLevelKey,
	logTypeKey,
	appEnvKey,
}

type Config struct {
	LogLevel slog.Level      `mapstructure:"log_level"`
	LogType  logging.LogType `mapstructure:"log_type"`
	AppEnv   string          `mapstructure:"app_env"`
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

	return config, nil
}

func setDefaults(v *viper.Viper) error {
	configparser.ParseFromKey(v, logLevelKey, logging.ParseLogLevel, slog.LevelInfo)
	configparser.ParseFromKey(v, logTypeKey, logging.ParseLogType, logging.LogTypeText)
	return nil
}

func NewLoggingConfig(config *Config) logging.Config {
	return logging.Config{
		Level: config.LogLevel,
		Type:  config.LogType,
	}
}
