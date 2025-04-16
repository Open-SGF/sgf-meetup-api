package api

import (
	"github.com/spf13/viper"
	"log/slog"
	"sgf-meetup-api/pkg/shared/configparser"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"strings"
)

const (
	logLevelKey        = "LOG_LEVEL"
	logTypeKey         = "LOG_TYPE"
	sentryDsnKey       = "SENTRY_DSN"
	dynamoDbEndpoint   = "DYNAMODB_ENDPOINT"
	awsRegion          = "AWS_REGION"
	awsAccessKey       = "AWS_ACCESS_KEY"
	awsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
)

var configKeys = []string{
	logLevelKey,
	logTypeKey,
	sentryDsnKey,
	dynamoDbEndpoint,
	awsRegion,
	awsAccessKey,
	awsSecretAccessKey,
}

type Config struct {
	LogLevel                      slog.Level      `mapstructure:"log_level"`
	LogType                       logging.LogType `mapstructure:"log_type"`
	SentryDsn                     string          `mapstructure:"sentry_dsn"`
	DynamoDbEndpoint              string          `mapstructure:"dynamodb_endpoint"`
	AwsRegion                     string          `mapstructure:"aws_region"`
	AwsAccessKey                  string          `mapstructure:"aws_access_key"`
	AwsSecretAccessKey            string          `mapstructure:"aws_secret_access_key"`
	EventsTableName               string
	GroupUrlNameDateTimeIndexName string
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

	config.EventsTableName = *db.EventsTableProps.TableName
	config.GroupUrlNameDateTimeIndexName = *db.GroupIdDateTimeIndex.IndexName

	return config, nil
}

func setDefaults(v *viper.Viper) {
	configparser.ParseFromKey(v, logLevelKey, logging.ParseLogLevel, slog.LevelInfo)
	configparser.ParseFromKey(v, logTypeKey, logging.ParseLogType, logging.LogTypeText)
	v.SetDefault(strings.ToLower(awsRegion), "us-east-2")
}

func getLoggingConfig(config *Config) logging.Config {
	return logging.Config{
		Level: config.LogLevel,
		Type:  config.LogType,
	}
}

func getDbConfig(config *Config) db.Config {
	return db.Config{
		Endpoint:        config.DynamoDbEndpoint,
		Region:          config.AwsRegion,
		AccessKey:       config.AwsAccessKey,
		SecretAccessKey: config.AwsSecretAccessKey,
	}
}
