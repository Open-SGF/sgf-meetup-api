package importer

import (
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"sgf-meetup-api/pkg/configparser"
	"strings"
)

const (
	logLevelKey                = "LOG_LEVEL"
	sentryDsnKey               = "SENTRY_DSN"
	meetupProxyFunctionNameKey = "MEETUP_PROXY_FUNCTION_NAME"
	eventsTableNameKey         = "EVENTS_TABLE_NAME"
	meetupGroupNamesKey        = "MEETUP_GROUP_NAMES"
	dynamoDbEndpoint           = "DYNAMODB_ENDPOINT"
	awsRegion                  = "AWS_REGION"
	awsAccessKey               = "AWS_ACCESS_KEY"
	awsSecretAccessKey         = "AWS_SECRET_ACCESS_KEY"
)

var keys = []string{
	strings.ToLower(logLevelKey),
	strings.ToLower(sentryDsnKey),
	strings.ToLower(meetupProxyFunctionNameKey),
	strings.ToLower(eventsTableNameKey),
	strings.ToLower(meetupGroupNamesKey),
	strings.ToLower(dynamoDbEndpoint),
	strings.ToLower(awsRegion),
	strings.ToLower(awsAccessKey),
	strings.ToLower(awsSecretAccessKey),
}

type Config struct {
	LogLevel                slog.Level `mapstructure:"log_level"`
	SentryDsn               string     `mapstructure:"sentry_dsn"`
	MeetupProxyFunctionName string     `mapstructure:"meetup_proxy_function_name"`
	EventsTableName         string     `mapstructure:"events_table_name"`
	MeetupGroupNames        []string   `mapstructure:"meetup_group_names"`
	DynamoDbEndpoint        string     `mapstructure:"dynamodb_endpoint"`
	AwsRegion               string     `mapstructure:"aws_region"`
	AwsAccessKey            string     `mapstructure:"aws_access_key"`
	AwsSecretAccessKey      string     `mapstructure:"aws_secret_access_key"`
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

	if err = validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults(v *viper.Viper) {
	configparser.ParseLogLevelFromKey(v, strings.ToLower(logLevelKey), slog.LevelInfo)
	v.SetDefault(strings.ToLower(meetupGroupNamesKey), []string{})
	v.SetDefault(strings.ToLower(awsRegion), "us-east-2")
}

func validateConfig(cfg *Config) error {
	var missing []string

	if cfg.MeetupProxyFunctionName == "" {
		missing = append(missing, meetupProxyFunctionNameKey)
	}
	if cfg.EventsTableName == "" {
		missing = append(missing, eventsTableNameKey)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %v", strings.Join(missing, ", "))
	}

	return nil
}
