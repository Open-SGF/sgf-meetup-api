package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/spf13/viper"
	"log"
	"sgf-meetup-api/src/importer"
	"strings"
)

var config *importer.Config

func init() {
	config = loadConfig()
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, event json.RawMessage) error {
	err := importer.Import(ctx, *config)

	return err
}

func loadConfig() *importer.Config {
	v := viper.New()

	v.SetDefault("meetup_group_names", []string{})

	v.AutomaticEnv()

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			log.Printf("Warning: error reading .env file: %v", err)
		}
	}

	var cfg importer.Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Printf("Unable to decode into struct: %v", err)
	}

	validateConfig(&cfg)

	return &cfg
}

func validateConfig(cfg *importer.Config) {
	var missing []string

	if cfg.GetTokenFunctionName == "" {
		missing = append(missing, "GET_TOKEN_FUNCTION_NAME")
	}
	if cfg.EventsTableName == "" {
		missing = append(missing, "EVENTS_TABLE_NAME")
	}
	if cfg.ImporterLogTableName == "" {
		missing = append(missing, "IMPORTER_LOG_TABLE_NAME")
	}

	if len(missing) > 0 {
		log.Fatalf("Missing required env vars: %v", strings.Join(missing, ", "))
	}
}
