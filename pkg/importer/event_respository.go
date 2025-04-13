package importer

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log/slog"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/models"
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
	db         *dynamodb.Client
	timeSource clock.TimeSource
	logger     *slog.Logger
}

func NewDynamoDBEventRepository(config DynamoDBEventRepositoryConfig, db *dynamodb.Client, timeSource clock.TimeSource, logger *slog.Logger) *DynamoDBEventRepository {
	return &DynamoDBEventRepository{
		config:     config,
		db:         db,
		timeSource: timeSource,
		logger:     logger,
	}
}

func (er *DynamoDBEventRepository) GetUpcomingEventsForGroup(ctx context.Context, group string) ([]models.MeetupEvent, error) {
	now := er.timeSource.Now().UTC().Format(time.RFC3339)

	var allEvents []models.MeetupEvent

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
	panic("implement me")
}

func (er *DynamoDBEventRepository) UpsertEvents(ctx context.Context, events []models.MeetupEvent) error {
	panic("implement me")
}
