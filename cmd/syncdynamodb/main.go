package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/spf13/viper"
	"log"
	"os"
	"sgf-meetup-api/src/db"
	"slices"
	"strings"
)

type Config struct {
	DynamoDbEndpoint   string `mapstructure:"dynamodb_endpoint"`
	Region             string `mapstructure:"region"`
	AwsAccessKey       string `mapstructure:"aws_access_key"`
	AwsSecretAccessKey string `mapstructure:"aws_secret_access_key"`
}

var config *Config

func init() {
	config = loadConfig()
}

func main() {
	ctx := context.Background()
	client, err := db.New(ctx, &db.Options{
		Endpoint:     config.DynamoDbEndpoint,
		Region:       config.Region,
		ClientKey:    config.AwsAccessKey,
		ClientSecret: config.AwsSecretAccessKey,
	})

	if err != nil {
		log.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	if err := syncDb(ctx, client); err != nil {
		log.Fatalf("Failed to sync database: %v", err)
	}
}

func syncDb(ctx context.Context, client *dynamodb.Client) error {
	return syncTables(ctx, client)
}

type CloudFormationTemplate struct {
	Resources map[string]Resource `json:"Resources"`
}

type Resource struct {
	Type       string                 `json:"Type"`
	Properties map[string]interface{} `json:"Properties"`
}

func syncTables(ctx context.Context, client *dynamodb.Client) error {
	listTablesOutput, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return fmt.Errorf("error listing tables: %w", err)
	}

	templateBytes, err := os.ReadFile("./cdk.out/SgfMeetupApiGo.template.json")
	if err != nil {
		return fmt.Errorf("error reading template file: %w", err)
	}

	var template CloudFormationTemplate
	if err := json.Unmarshal(templateBytes, &template); err != nil {
		return fmt.Errorf("error parsing template JSON: %w", err)
	}

	for resourceKey, resourceValue := range template.Resources {
		if resourceValue.Type != "AWS::DynamoDB::Table" {
			continue
		}

		tableNameInterface, ok := resourceValue.Properties["TableName"]
		if !ok {
			continue
		}

		tableName, ok := tableNameInterface.(string)
		if !ok {
			continue
		}

		tableExists := slices.Contains(listTablesOutput.TableNames, tableName)

		if tableExists {
			log.Printf("Table %s already exists, skipping", tableName)
			continue
		}

		// Convert properties to CreateTableInput
		createTableInput, err := convertToCreateTableInput(resourceValue.Properties)
		if err != nil {
			log.Printf("Error when converting properties for table resource %s: %v", resourceKey, err)
			continue
		}

		// Create the table
		_, err = client.CreateTable(ctx, createTableInput)
		if err != nil {
			log.Printf("Error when creating table resource %s: %v", resourceKey, err)
			continue
		}

		log.Printf("Table %s created", tableName)
	}

	return nil
}

func convertToCreateTableInput(properties map[string]interface{}) (*dynamodb.CreateTableInput, error) {
	propBytes, err := json.Marshal(properties)
	if err != nil {
		return nil, fmt.Errorf("error marshaling properties: %w", err)
	}

	var createTableInput dynamodb.CreateTableInput
	if err := json.Unmarshal(propBytes, &createTableInput); err != nil {
		return nil, fmt.Errorf("error unmarshaling to CreateTableInput: %w", err)
	}

	return &createTableInput, nil
}

func loadConfig() *Config {
	v := viper.New()

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

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Printf("Unable to decode into struct: %v", err)
	}

	validateConfig(&cfg)

	return &cfg
}

func validateConfig(cfg *Config) {
	var missing []string

	if cfg.DynamoDbEndpoint == "" {
		missing = append(missing, "DYNAMODB_ENDPOINT")
	}
	if cfg.Region == "" {
		missing = append(missing, "REGION")
	}
	if cfg.AwsAccessKey == "" {
		missing = append(missing, "AWS_ACCESS_KEY")
	}
	if cfg.AwsSecretAccessKey == "" {
		missing = append(missing, "AWS_SECRET_ACCESS_KEY")
	}

	if len(missing) > 0 {
		log.Fatalf("Missing required env vars: %v", strings.Join(missing, ", "))
	}
}
