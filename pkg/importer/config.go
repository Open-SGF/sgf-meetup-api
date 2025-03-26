package importer

import (
	"fmt"
	"github.com/spf13/viper"
	"sgf-meetup-api/pkg/configparser"
	"strings"
)

const (
	meetupProxyFunctionName = "MEETUP_PROXY_FUNCTION_NAME"
	eventsTableNameKey      = "EVENTS_TABLE_NAME"
	importerLogTableNameKey = "IMPORTER_LOG_TABLE_NAME"
	meetupGroupNamesKey     = "MEETUP_GROUP_NAMES"
)

var keys = []string{
	strings.ToLower(meetupProxyFunctionName),
	strings.ToLower(eventsTableNameKey),
	strings.ToLower(importerLogTableNameKey),
	strings.ToLower(meetupGroupNamesKey),
}

type Config struct {
	MeetupProxyFunctionName string   `mapstructure:"meetup_proxy_function_name"`
	EventsTableName         string   `mapstructure:"events_table_name"`
	ImporterLogTableName    string   `mapstructure:"importer_log_table_name"`
	MeetupGroupNames        []string `mapstructure:"meetup_group_names"`
	//DynamoDbEndpoint        string   `mapstructure:"dynamodb_endpoint"`
	//AwsRegion               string   `mapstructure:"aws_region"`
	//AwsAccessKey            string   `mapstructure:"aws_access_key"`
	//AwsSecretAccessKey      string   `mapstructure:"aws_secret_access_key"`
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
	v.SetDefault(strings.ToLower(meetupGroupNamesKey), []string{})
}

func validateConfig(cfg *Config) error {
	var missing []string

	if cfg.MeetupProxyFunctionName == "" {
		missing = append(missing, meetupProxyFunctionName)
	}
	if cfg.EventsTableName == "" {
		missing = append(missing, eventsTableNameKey)
	}
	if cfg.ImporterLogTableName == "" {
		missing = append(missing, importerLogTableNameKey)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %v", strings.Join(missing, ", "))
	}

	return nil
}
