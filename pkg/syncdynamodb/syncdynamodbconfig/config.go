package syncdynamodbconfig

import (
	"context"
	"github.com/spf13/viper"
	"log/slog"
	"sgf-meetup-api/pkg/shared/configparser"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"strings"
)

const (
	logLevelKey           = "LOG_LEVEL"
	logTypeKey            = "LOG_TYPE"
	dynamoDBEndpointKey   = "DYNAMODB_ENDPOINT"
	awsRegionKey          = "AWS_REGION"
	awsAccessKeyKey       = "AWS_ACCESS_KEY"
	awsSecretAccessKeyKey = "AWS_SECRET_ACCESS_KEY"
)

var configKeys = []string{
	logLevelKey,
	logTypeKey,
	dynamoDBEndpointKey,
	awsRegionKey,
	awsAccessKeyKey,
	awsSecretAccessKeyKey,
}

type Config struct {
	LogLevel           slog.Level      `mapstructure:"log_level"`
	LogType            logging.LogType `mapstructure:"log_type"`
	DynamoDbEndpoint   string          `mapstructure:"dynamodb_endpoint"`
	AwsRegion          string          `mapstructure:"aws_region"`
	AwsAccessKey       string          `mapstructure:"aws_access_key"`
	AwsSecretAccessKey string          `mapstructure:"aws_secret_access_key"`
}

func NewConfig(ctx context.Context) (*Config, error) {
	return NewConfigFromEnvFile(ctx, ".", ".env")
}

func NewConfigFromEnvFile(ctx context.Context, path, filename string) (*Config, error) {
	config, err := configparser.Parse[Config](ctx, configparser.ParseOptions{
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
	v.SetDefault(strings.ToLower(awsRegionKey), "us-east-2")
	return nil
}

func NewLoggingConfig(config *Config) logging.Config {
	return logging.Config{
		Level: config.LogLevel,
		Type:  config.LogType,
	}
}

func NewDBConfig(config *Config) db.Config {
	return db.Config{
		Endpoint:        config.DynamoDbEndpoint,
		Region:          config.AwsRegion,
		AccessKey:       config.AwsAccessKey,
		SecretAccessKey: config.AwsSecretAccessKey,
	}
}
