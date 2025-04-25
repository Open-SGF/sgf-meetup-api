package importerconfig

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"sgf-meetup-api/pkg/shared/appconfig"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"strings"
)

const (
	logLevelKey                 = "LOG_LEVEL"
	logTypeKey                  = "LOG_TYPE"
	sentryDsnKey                = "SENTRY_DSN"
	meetupGroupNamesKey         = "MEETUP_GROUP_NAMES"
	dynamoDbEndpointKey         = "DYNAMODB_ENDPOINT"
	awsRegionKey                = "AWS_REGION"
	awsAccessKeyKey             = "AWS_ACCESS_KEY"
	awsSecretAccessKeyKey       = "AWS_SECRET_ACCESS_KEY"
	proxyFunctionNameKey        = "MEETUP_PROXY_FUNCTION_NAME"
	archivedEventsTableNameKey  = "ARCHIVED_EVENTS_TABLE_NAME"
	eventsTableNameKey          = "EVENTS_TABLE_NAME"
	groupIDDateTimeIndexNameKey = "GROUP_ID_DATE_TIME_INDEX_NAME"
)

var configKeys = []string{
	logLevelKey,
	logTypeKey,
	sentryDsnKey,
	meetupGroupNamesKey,
	dynamoDbEndpointKey,
	awsRegionKey,
	awsAccessKeyKey,
	awsSecretAccessKeyKey,
	proxyFunctionNameKey,
	archivedEventsTableNameKey,
	eventsTableNameKey,
	groupIDDateTimeIndexNameKey,
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

	if err = config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults(v *viper.Viper) error {
	appconfig.ParseFromKey(v, logLevelKey, logging.ParseLogLevel, slog.LevelInfo)
	appconfig.ParseFromKey(v, logTypeKey, logging.ParseLogType, logging.LogTypeText)
	v.SetDefault(strings.ToLower(meetupGroupNamesKey), []string{})
	v.SetDefault(strings.ToLower(awsRegionKey), "us-east-2")
	return nil
}

func (config *Config) validate() error {
	var missing []string

	if config.ProxyFunctionName == "" {
		missing = append(missing, proxyFunctionNameKey)
	}
	if config.EventsTableName == "" {
		missing = append(missing, eventsTableNameKey)
	}
	if config.ArchivedEventsTableName == "" {
		missing = append(missing, archivedEventsTableNameKey)
	}
	if config.GroupIDDateTimeIndexName == "" {
		missing = append(missing, groupIDDateTimeIndexNameKey)
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
