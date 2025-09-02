package db

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"reflect"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/logging"
)

type TestDB struct {
	Client    *Client
	Container *tcdynamodb.DynamoDBContainer
	logger    *slog.Logger
}

func NewTestDB(ctx context.Context) (*TestDB, error) {
	testDB, err := NewTestDBWithoutMigrations(ctx)
	if err != nil {
		return nil, err
	}

	if err = SyncTables(ctx, testDB.logger, testDB.Client, "", infra.Tables); err != nil {
		return nil, err
	}

	return testDB, nil
}

func NewTestDBWithoutMigrations(ctx context.Context) (*TestDB, error) {
	ctr, err := tcdynamodb.Run(
		ctx,
		"amazon/dynamodb-local:3.0.0",
		tcdynamodb.WithDisableTelemetry(),
	)
	if err != nil {
		return nil, err
	}

	connectionString, err := ctr.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	dbOptions := Config{
		Endpoint:        "http://" + connectionString,
		Region:          "us-east-2",
		SecretAccessKey: "test",
		AccessKey:       "test",
	}

	logger := logging.DefaultLogger(
		ctx,
		logging.Config{Level: slog.LevelError, Type: logging.LogTypeText},
	)

	client, err := NewClient(ctx, dbOptions, nil, logger)
	if err != nil {
		return nil, err
	}

	testDB := TestDB{
		Client:    client,
		Container: ctr,
		logger:    logger,
	}

	return &testDB, nil
}

func (ctr *TestDB) Close() {
	if ctrErr := testcontainers.TerminateContainer(ctr.Container); ctrErr != nil {
		log.Printf("failed to terminate Container: %s", ctrErr)
	}
}

func (ctr *TestDB) Reset(ctx context.Context) error {
	for _, table := range infra.Tables {
		var deleteRequests []types.WriteRequest

		paginator := dynamodb.NewScanPaginator(ctr.Client, &dynamodb.ScanInput{
			TableName:            table.TableName,
			ProjectionExpression: table.PartitionKey.Name,
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return fmt.Errorf("scan failed for %s: %w", *table.TableName, err)
			}

			for _, item := range page.Items {
				deleteRequests = append(deleteRequests, types.WriteRequest{
					DeleteRequest: &types.DeleteRequest{
						Key: map[string]types.AttributeValue{
							*table.PartitionKey.Name: item[*table.PartitionKey.Name],
						},
					},
				})
			}
		}

		for chunk := range slices.Chunk(deleteRequests, MaxBatchSize) {
			input := &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{
					*table.TableName: chunk,
				},
			}

			for {
				resp, err := ctr.Client.BatchWriteItem(ctx, input)
				if err != nil {
					return fmt.Errorf("batch delete failed for %s: %w", *table.TableName, err)
				}

				if len(resp.UnprocessedItems) == 0 {
					break
				}

				input.RequestItems = resp.UnprocessedItems
			}
		}
	}

	return nil
}

func (ctr *TestDB) InsertTestItems(ctx context.Context, tableName string, testItems any) {
	v := reflect.ValueOf(testItems)
	if v.Kind() != reflect.Slice {
		panic(fmt.Sprintf("expected slice, got %T", testItems))
	}

	items := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		items[i] = v.Index(i).Interface()
	}

	if len(items) == 0 {
		return
	}

	for chunk := range slices.Chunk(items, MaxBatchSize) {
		writeRequests := make([]types.WriteRequest, 0, len(chunk))

		for _, event := range chunk {
			av, err := attributevalue.MarshalMap(event)
			if err != nil {
				panic("error marshaling item")
			}

			writeRequests = append(writeRequests, types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: av,
				},
			})
		}

		_, err := ctr.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				tableName: writeRequests,
			},
		})
		if err != nil {
			panic("error writing item")
		}
	}
}

func (ctr *TestDB) CheckItemExists(ctx context.Context, tableName, keyName, key string) bool {
	resp, err := ctr.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			keyName: &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		panic("error while checking if item exists")
	}

	return len(resp.Item) > 0
}

func (ctr *TestDB) GetItemCount(ctx context.Context, tableName string) int {
	scanInput := &dynamodb.ScanInput{TableName: aws.String(tableName)}
	paginator := dynamodb.NewScanPaginator(ctr.Client, scanInput)

	count := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			panic("error getting item count")
		}

		count += len(page.Items)
	}

	return count
}
