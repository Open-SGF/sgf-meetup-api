package db

import (
	"context"
	"errors"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log/slog"
)

func SyncTables(ctx context.Context, logger *slog.Logger, client *Client, tables []DynamoDbProps) error {

	for _, tableProps := range tables {
		tableName := *tableProps.TableName

		exists, err := tableExists(ctx, client, logger, tableName)
		if err != nil {
			return err
		}

		if exists {
			logger.Info("Table already exists", "tableName", tableName)
			continue
		}

		if err := createTable(ctx, client, tableProps); err != nil {
			return err
		}

		logger.Info("Created table", "tableName", tableName)
	}

	return nil
}

func tableExists(ctx context.Context, client *Client, logger *slog.Logger, tableName string) (bool, error) {
	logger.Info("checking table", "tableName", tableName)
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

func createTable(ctx context.Context, client *Client, tableProps DynamoDbProps) error {
	attrMap := make(map[string]types.ScalarAttributeType)

	partitionKeyName := *tableProps.PartitionKey.Name
	attrMap[partitionKeyName] = convertAttributeType(tableProps.PartitionKey.Type)

	keySchema := []types.KeySchemaElement{
		{
			AttributeName: aws.String(partitionKeyName),
			KeyType:       types.KeyTypeHash,
		},
	}

	if tableProps.SortKey != nil {
		sortKeyName := *tableProps.SortKey.Name
		attrMap[sortKeyName] = convertAttributeType(tableProps.SortKey.Type)

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
		TableName:            tableProps.TableName,
		AttributeDefinitions: attrDefs,
		KeySchema:            keySchema,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(floatToIntWithFallback(tableProps.ReadCapacity, 5)),
			WriteCapacityUnits: aws.Int64(floatToIntWithFallback(tableProps.WriteCapacity, 5)),
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
