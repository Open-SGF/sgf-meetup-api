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

type EventDBRepository interface {
	GetUpcomingEventsForGroup(ctx context.Context, group string) ([]models.MeetupEvent, error)
}

type EventDBRepositoryConfig struct {
	TableName          string
	GroupDateIndexName string
}

func NewEventDBRepositoryConfig(config *Config) EventDBRepositoryConfig {
	return EventDBRepositoryConfig{
		TableName:          config.EventsTableName,
		GroupDateIndexName: config.GroupUrlNameDateTimeIndexName,
	}
}

type eventDBRepository struct {
	config     EventDBRepositoryConfig
	db         *dynamodb.Client
	timeSource clock.TimeSource
	logger     *slog.Logger
}

func NewEventDBRepository(config EventDBRepositoryConfig, db *dynamodb.Client, timeSource clock.TimeSource, logger *slog.Logger) EventDBRepository {
	return &eventDBRepository{
		config:     config,
		db:         db,
		timeSource: timeSource,
		logger:     logger,
	}
}

func (er *eventDBRepository) GetUpcomingEventsForGroup(ctx context.Context, group string) ([]models.MeetupEvent, error) {
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
		TableName:                 aws.String(er.config.TableName),
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
