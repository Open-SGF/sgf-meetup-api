package importer

import (
	"github.com/spf13/viper"
	"log/slog"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/configparser"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"strings"
)

const (
	logLevelKey         = "LOG_LEVEL"
	logTypeKey          = "LOG_TYPE"
	sentryDsnKey        = "SENTRY_DSN"
	meetupGroupNamesKey = "MEETUP_GROUP_NAMES"
	dynamoDbEndpoint    = "DYNAMODB_ENDPOINT"
	awsRegion           = "AWS_REGION"
	awsAccessKey        = "AWS_ACCESS_KEY"
	awsSecretAccessKey  = "AWS_SECRET_ACCESS_KEY"
)

var keys = []string{
	logLevelKey,
	logTypeKey,
	sentryDsnKey,
	meetupGroupNamesKey,
	dynamoDbEndpoint,
	awsRegion,
	awsAccessKey,
	awsSecretAccessKey,
}

type Config struct {
	LogLevel                      slog.Level      `mapstructure:"log_level"`
	LogType                       logging.LogType `mapstructure:"log_type"`
	SentryDsn                     string          `mapstructure:"sentry_dsn"`
	MeetupGroupNames              []string        `mapstructure:"meetup_group_names"`
	DynamoDbEndpoint              string          `mapstructure:"dynamodb_endpoint"`
	AwsRegion                     string          `mapstructure:"aws_region"`
	AwsAccessKey                  string          `mapstructure:"aws_access_key"`
	AwsSecretAccessKey            string          `mapstructure:"aws_secret_access_key"`
	ProxyFunctionName             string
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
		Keys:        keys,
		SetDefaults: setDefaults,
	})

	if err != nil {
		return nil, err
	}

	config.ProxyFunctionName = *infra.MeetupProxyFunctionName
	config.EventsTableName = *infra.EventsTableProps.TableName
	config.GroupUrlNameDateTimeIndexName = *infra.GroupUrlNameDateTimeIndex.IndexName

	return config, nil
}

func setDefaults(v *viper.Viper) {
	configparser.ParseFromKey(v, logLevelKey, logging.ParseLogLevel, slog.LevelInfo)
	configparser.ParseFromKey(v, logTypeKey, logging.ParseLogType, logging.LogTypeText)
	v.SetDefault(strings.ToLower(meetupGroupNamesKey), []string{})
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
