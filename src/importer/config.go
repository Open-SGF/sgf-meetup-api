package importer

import (
	"errors"
	"github.com/spf13/viper"
	"log"
	"strings"
)

const (
	meetupTokenFunctionNameKey = "MEETUP_TOKEN_FUNCTION_NAME"
	eventsTableNameKey         = "EVENTS_TABLE_NAME"
	importerLogTableNameKey    = "IMPORTER_LOG_TABLE_NAME"
	meetupGroupNamesKey        = "MEETUP_GROUP_NAMES"
)

type Config struct {
	MeetupTokenFunctionName string   `mapstructure:"meetup_token_function_name"`
	EventsTableName         string   `mapstructure:"events_table_name"`
	ImporterLogTableName    string   `mapstructure:"importer_log_table_name"`
	MeetupGroupNames        []string `mapstructure:"meetup_group_names"`
}

func LoadConfig() *Config {
	v := viper.New()

	v.SetDefault(strings.ToLower(meetupTokenFunctionNameKey), "")
	v.SetDefault(strings.ToLower(eventsTableNameKey), "")
	v.SetDefault(strings.ToLower(importerLogTableNameKey), "")
	v.SetDefault(strings.ToLower(meetupGroupNamesKey), []string{})

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			log.Printf("Warning: error reading .env file: %v", err)
		}
	}

	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Printf("Unable to decode into struct: %v", err)
	}

	validateConfig(&cfg)

	return &cfg
}

func validateConfig(cfg *Config) {
	var missing []string

	if cfg.MeetupTokenFunctionName == "" {
		missing = append(missing, meetupTokenFunctionNameKey)
	}
	if cfg.EventsTableName == "" {
		missing = append(missing, eventsTableNameKey)
	}
	if cfg.ImporterLogTableName == "" {
		missing = append(missing, importerLogTableNameKey)
	}

	if len(missing) > 0 {
		log.Fatalf("Missing required env vars: %v", strings.Join(missing, ", "))
	}
}
