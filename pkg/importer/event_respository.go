package importer

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log/slog"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/models"
	"slices"
	"time"
)

type EventRepository interface {
	GetUpcomingEventsForGroup(ctx context.Context, group string) ([]models.MeetupEvent, error)
	ArchiveEvents(ctx context.Context, eventIds []string) error
	UpsertEvents(ctx context.Context, events []models.MeetupEvent) error
}

type DynamoDBEventRepositoryConfig struct {
	EventsTableName         string
	ArchivedEventsTableName string
	GroupDateIndexName      string
}

func NewDynamoDBEventRepositoryConfig(config *Config) DynamoDBEventRepositoryConfig {
	return DynamoDBEventRepositoryConfig{
		EventsTableName:         config.EventsTableName,
		ArchivedEventsTableName: config.ArchivedEventsTableName,
		GroupDateIndexName:      config.GroupUrlNameDateTimeIndexName,
	}
}

type DynamoDBEventRepository struct {
	config     DynamoDBEventRepositoryConfig
	db         *db.Client
	timeSource clock.TimeSource
	logger     *slog.Logger
}

func NewDynamoDBEventRepository(config DynamoDBEventRepositoryConfig, db *db.Client, timeSource clock.TimeSource, logger *slog.Logger) *DynamoDBEventRepository {
	return &DynamoDBEventRepository{
		config:     config,
		db:         db,
		timeSource: timeSource,
		logger:     logger,
	}
}

func (er *DynamoDBEventRepository) GetUpcomingEventsForGroup(ctx context.Context, group string) ([]models.MeetupEvent, error) {
	now := er.timeSource.Now().UTC().Format(time.RFC3339)

	keyCond := expression.Key("groupId").
		Equal(expression.Value(group)).
		And(expression.Key("dateTime").
			GreaterThan(expression.Value(now)))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()

	if err != nil {
		return nil, err
	}

	paginator := dynamodb.NewQueryPaginator(er.db, &dynamodb.QueryInput{
		TableName:                 aws.String(er.config.EventsTableName),
		IndexName:                 aws.String(er.config.GroupDateIndexName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	var allEvents []models.MeetupEvent

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		var events []models.MeetupEvent
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &events); err != nil {
			return nil, err
		}
		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

func (er *DynamoDBEventRepository) ArchiveEvents(ctx context.Context, eventIds []string) error {
	if len(eventIds) == 0 {
		return nil
	}

	for chunk := range slices.Chunk(eventIds, db.MaxBatchSize) {
		items, err := er.getItems(ctx, chunk)
		if err != nil {
			return err
		}

		if len(items) <= 0 {
			continue
		}

		if err := er.writeToArchive(ctx, items); err != nil {
			return err
		}

		if err := er.deleteIds(ctx, chunk); err != nil {
			return fmt.Errorf("delete chunk: %w", err)
		}
	}
	return nil
}

func (er *DynamoDBEventRepository) UpsertEvents(ctx context.Context, events []models.MeetupEvent) error {
	return er.upsertEventsToTable(ctx, events, er.config.EventsTableName)
}

func (er *DynamoDBEventRepository) upsertEventsToTable(ctx context.Context, events []models.MeetupEvent, table string) error {
	if len(events) == 0 {
		return nil
	}

	for chunk := range slices.Chunk(events, db.MaxBatchSize) {
		writeRequests := make([]types.WriteRequest, 0, len(chunk))

		for _, event := range chunk {
			av, err := attributevalue.MarshalMap(event)

			if err != nil {
				return err
			}

			writeRequests = append(writeRequests, types.WriteRequest{
				PutRequest: &types.PutRequest{Item: av},
			})
		}

		_, err := er.db.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				table: writeRequests,
			},
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (er *DynamoDBEventRepository) getItems(ctx context.Context, ids []string) ([]map[string]types.AttributeValue, error) {
	keys := make([]map[string]types.AttributeValue, len(ids))
	for i, id := range ids {
		keys[i] = er.createKey(id)
	}

	res, err := er.db.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			er.config.EventsTableName: {Keys: keys},
		},
	})
	if err != nil {
		return nil, err
	}
	return res.Responses[er.config.EventsTableName], nil
}

func (er *DynamoDBEventRepository) writeToArchive(ctx context.Context, items []map[string]types.AttributeValue) error {
	writes := make([]types.WriteRequest, len(items))
	for i, item := range items {
		writes[i] = types.WriteRequest{PutRequest: &types.PutRequest{Item: item}}
	}

	_, err := er.db.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{er.config.ArchivedEventsTableName: writes},
	})

	return err
}

func (er *DynamoDBEventRepository) deleteIds(ctx context.Context, ids []string) error {
	deletes := make([]types.WriteRequest, len(ids))
	for i, id := range ids {
		deletes[i] = types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{Key: er.createKey(id)},
		}
	}

	_, err := er.db.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{er.config.EventsTableName: deletes},
	})

	return err
}

func (er *DynamoDBEventRepository) createKey(id string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}}
}
