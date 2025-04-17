package api

import (
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"sgf-meetup-api/pkg/shared/configparser"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"strings"
)

const (
	logLevelKey              = "LOG_LEVEL"
	logTypeKey               = "LOG_TYPE"
	sentryDSNKey             = "SENTRY_DSN"
	dynamoDbEndpoint         = "DYNAMODB_ENDPOINT"
	awsRegion                = "AWS_REGION"
	awsAccessKey             = "AWS_ACCESS_KEY"
	awsSecretAccessKey       = "AWS_SECRET_ACCESS_KEY"
	eventsTableName          = "EVENTS_TABLE_NAME"
	apiUsersTableName        = "API_USERS_TABLE_NAME"
	groupIDDateTimeIndexName = "GROUP_ID_DATE_TIME_INDEX_NAME"
)

var configKeys = []string{
	logLevelKey,
	logTypeKey,
	sentryDSNKey,
	dynamoDbEndpoint,
	awsRegion,
	awsAccessKey,
	awsSecretAccessKey,
	eventsTableName,
	apiUsersTableName,
	groupIDDateTimeIndexName,
}

type Config struct {
	LogLevel                 slog.Level      `mapstructure:"log_level"`
	LogType                  logging.LogType `mapstructure:"log_type"`
	SentryDsn                string          `mapstructure:"sentry_dsn"`
	DynamoDbEndpoint         string          `mapstructure:"dynamodb_endpoint"`
	AwsRegion                string          `mapstructure:"aws_region"`
	AwsAccessKey             string          `mapstructure:"aws_access_key"`
	AwsSecretAccessKey       string          `mapstructure:"aws_secret_access_key"`
	EventsTableName          string          `mapstructure:"events_table_name"`
	ApiUsersTableName        string          `mapstructure:"api_users_table_name"`
	GroupIDDateTimeIndexName string          `mapstructure:"group_id_date_time_index_name"`
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

func setDefaults(v *viper.Viper) {
	configparser.ParseFromKey(v, logLevelKey, logging.ParseLogLevel, slog.LevelInfo)
	configparser.ParseFromKey(v, logTypeKey, logging.ParseLogType, logging.LogTypeText)
	v.SetDefault(strings.ToLower(awsRegion), "us-east-2")
}

func (config *Config) validate() error {
	var missing []string

	if config.EventsTableName == "" {
		missing = append(missing, eventsTableName)
	}
	if config.ApiUsersTableName == "" {
		missing = append(missing, apiUsersTableName)
	}
	if config.GroupIDDateTimeIndexName == "" {
		missing = append(missing, groupIDDateTimeIndexName)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %v", strings.Join(missing, ", "))
	}

	return nil
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
