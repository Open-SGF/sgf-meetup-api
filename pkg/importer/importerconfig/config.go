package importerconfig

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
	sentryDsnKey             = "SENTRY_DSN"
	meetupGroupNamesKey      = "MEETUP_GROUP_NAMES"
	dynamoDbEndpoint         = "DYNAMODB_ENDPOINT"
	awsRegion                = "AWS_REGION"
	awsAccessKey             = "AWS_ACCESS_KEY"
	awsSecretAccessKey       = "AWS_SECRET_ACCESS_KEY"
	proxyFunctionName        = "MEETUP_PROXY_FUNCTION_NAME"
	archivedEventsTableName  = "ARCHIVED_EVENTS_TABLE_NAME"
	eventsTableName          = "EVENTS_TABLE_NAME"
	groupIDDateTimeIndexName = "GROUP_ID_DATE_TIME_INDEX_NAME"
)

var configKeys = []string{
	logLevelKey,
	logTypeKey,
	sentryDsnKey,
	meetupGroupNamesKey,
	dynamoDbEndpoint,
	awsRegion,
	awsAccessKey,
	awsSecretAccessKey,
	proxyFunctionName,
	archivedEventsTableName,
	eventsTableName,
	groupIDDateTimeIndexName,
}

type Config struct {
	LogLevel                 slog.Level      `mapstructure:"log_level"`
	LogType                  logging.LogType `mapstructure:"log_type"`
	SentryDsn                string          `mapstructure:"sentry_dsn"`
	MeetupGroupNames         []string        `mapstructure:"meetup_group_names"`
	DynamoDbEndpoint         string          `mapstructure:"dynamodb_endpoint"`
	AwsRegion                string          `mapstructure:"aws_region"`
	AwsAccessKey             string          `mapstructure:"aws_access_key"`
	AwsSecretAccessKey       string          `mapstructure:"aws_secret_access_key"`
	ProxyFunctionName        string          `mapstructure:"meetup_proxy_function_name"`
	ArchivedEventsTableName  string          `mapstructure:"archived_events_table_name"`
	EventsTableName          string          `mapstructure:"events_table_name"`
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
	v.SetDefault(strings.ToLower(meetupGroupNamesKey), []string{})
	v.SetDefault(strings.ToLower(awsRegion), "us-east-2")
}

func (config *Config) validate() error {
	var missing []string

	if config.ProxyFunctionName == "" {
		missing = append(missing, proxyFunctionName)
	}
	if config.EventsTableName == "" {
		missing = append(missing, eventsTableName)
	}
	if config.ArchivedEventsTableName == "" {
		missing = append(missing, archivedEventsTableName)
	}
	if config.GroupIDDateTimeIndexName == "" {
		missing = append(missing, groupIDDateTimeIndexName)
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

func NewDBConfig(config *Config) db.Config {
	return db.Config{
		Endpoint:        config.DynamoDbEndpoint,
		Region:          config.AwsRegion,
		AccessKey:       config.AwsAccessKey,
		SecretAccessKey: config.AwsSecretAccessKey,
	}
}
