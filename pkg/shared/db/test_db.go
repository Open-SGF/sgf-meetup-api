package db

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"log"
	"log/slog"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/logging"
	"slices"
)

type TestDB struct {
	Client    *Client
	Container *tcdynamodb.DynamoDBContainer
	logger    *slog.Logger
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

func NewTestDB(ctx context.Context) (*TestDB, error) {
	testDB, err := NewTestDBWithoutMigrations(ctx)

	if err != nil {
		return nil, err
	}

	if err = SyncTables(ctx, testDB.logger, testDB.Client, infra.Tables); err != nil {
		return nil, err
	}

	return testDB, nil
}

func NewTestDBWithoutMigrations(ctx context.Context) (*TestDB, error) {
	ctr, err := tcdynamodb.Run(
		ctx,
		"amazon/dynamodb-local:2.6.0",
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

	logger := logging.DefaultLogger(logging.Config{Level: slog.LevelError, Type: logging.LogTypeText})

	client, err := NewClient(ctx, dbOptions, logger)

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
