package main

import (
	"context"
	"errors"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/spf13/viper"
	"log"
	"sgf-meetup-api/src/db"
	"sgf-meetup-api/src/infra"
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

func syncTables(ctx context.Context, client *dynamodb.Client) error {
	tables := []infra.DynamoDbProps{infra.EventsTableProps, infra.ImportLogsTableProps}

	for _, tableProps := range tables {
		tableName := *tableProps.TableProps.TableName

		exists, err := tableExists(ctx, client, tableName)
		if err != nil {
			return err
		}

		if exists {
			log.Printf("Table already exists: %s", tableName)
			continue
		}

		if err := createTable(ctx, client, tableProps); err != nil {
			return err
		}

		log.Printf("Created table: %s", tableName)
	}

	return nil
}

func tableExists(ctx context.Context, client *dynamodb.Client, tableName string) (bool, error) {
	log.Println("checking table", tableName)
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func createTable(ctx context.Context, client *dynamodb.Client, tableProps infra.DynamoDbProps) error {
	attrMap := make(map[string]types.ScalarAttributeType)

	partitionKeyName := *tableProps.TableProps.PartitionKey.Name
	attrMap[partitionKeyName] = convertAttributeType(tableProps.TableProps.PartitionKey.Type)

	keySchema := []types.KeySchemaElement{
		{
			AttributeName: aws.String(partitionKeyName),
			KeyType:       types.KeyTypeHash,
		},
	}

	if tableProps.TableProps.SortKey != nil {
		sortKeyName := *tableProps.TableProps.SortKey.Name
		attrMap[sortKeyName] = convertAttributeType(tableProps.TableProps.SortKey.Type)

		keySchema = append(keySchema, types.KeySchemaElement{
			AttributeName: aws.String(sortKeyName),
			KeyType:       types.KeyTypeRange,
		})
	}

	var gsis []types.GlobalSecondaryIndex
	if len(tableProps.GlobalSecondaryIndexes) > 0 {
		for _, gsi := range tableProps.GlobalSecondaryIndexes {
			// Add GSI partition key to attribute definitions
			gsiPartitionKeyName := *gsi.PartitionKey.Name
			attrMap[gsiPartitionKeyName] = convertAttributeType(gsi.PartitionKey.Type)

			gsiKeySchema := []types.KeySchemaElement{
				{
					AttributeName: aws.String(gsiPartitionKeyName),
					KeyType:       types.KeyTypeHash,
				},
			}

			if gsi.SortKey != nil {
				gsiSortKeyName := *gsi.SortKey.Name
				attrMap[gsiSortKeyName] = convertAttributeType(gsi.SortKey.Type)

				gsiKeySchema = append(gsiKeySchema, types.KeySchemaElement{
					AttributeName: aws.String(gsiSortKeyName),
					KeyType:       types.KeyTypeRange,
				})
			}

			gsis = append(gsis, types.GlobalSecondaryIndex{
				IndexName: gsi.IndexName,
				KeySchema: gsiKeySchema,
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(floatToIntWithFallback(gsi.ReadCapacity, 5)),
					WriteCapacityUnits: aws.Int64(floatToIntWithFallback(gsi.WriteCapacity, 5)),
				},
			})
		}
	}

	var attrDefs []types.AttributeDefinition
	for name, attrType := range attrMap {
		attrDefs = append(attrDefs, types.AttributeDefinition{
			AttributeName: aws.String(name),
			AttributeType: attrType,
		})
	}

	input := &dynamodb.CreateTableInput{
		TableName:            tableProps.TableProps.TableName,
		AttributeDefinitions: attrDefs,
		KeySchema:            keySchema,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(floatToIntWithFallback(tableProps.TableProps.ReadCapacity, 5)),
			WriteCapacityUnits: aws.Int64(floatToIntWithFallback(tableProps.TableProps.WriteCapacity, 5)),
		},
	}

	if len(gsis) > 0 {
		input.GlobalSecondaryIndexes = gsis
	}

	_, err := client.CreateTable(ctx, input)
	return err
}

func floatToIntWithFallback(float *float64, fallback int64) int64 {
	value := fallback

	if float != nil {
		value = int64(*float)
	}

	return value
}

func convertAttributeType(cdkAttrType awsdynamodb.AttributeType) types.ScalarAttributeType {
	switch cdkAttrType {
	case awsdynamodb.AttributeType_STRING:
		return types.ScalarAttributeTypeS
	case awsdynamodb.AttributeType_NUMBER:
		return types.ScalarAttributeTypeN
	case awsdynamodb.AttributeType_BINARY:
		return types.ScalarAttributeTypeB
	default:
		return types.ScalarAttributeTypeS
	}
}

func loadConfig() *Config {
	v := viper.New()

	v.SetDefault("dynamodb_endpoint", "")
	v.SetDefault("region", "us-east-2")
	v.SetDefault("aws_access_key", "")
	v.SetDefault("aws_secret_access_key", "")

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
