package groupevents

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/models"
	"time"
)

type PaginatedEventsFilters struct {
	Before *time.Time
	After  *time.Time
	Cursor string
	Limit  *int
}

type GroupEventRepository interface {
	PaginatedEvents(ctx context.Context, groupID string, filters PaginatedEventsFilters) ([]models.MeetupEvent, *PaginatedEventsFilters, error)
}

type DynamoDBGroupEventRepositoryConfig struct {
	EventsTableName    string
	GroupDateIndexName string
}

func NewDynamoDBGroupEventRepositoryConfig(config *apiconfig.Config) DynamoDBGroupEventRepositoryConfig {
	return DynamoDBGroupEventRepositoryConfig{
		EventsTableName:    config.EventsTableName,
		GroupDateIndexName: config.GroupIDDateTimeIndexName,
	}
}

type DynamoDBGroupEventRepository struct {
	config     DynamoDBGroupEventRepositoryConfig
	timeSource clock.TimeSource
	db         *db.Client
}

func NewDynamoDBGroupEventRepository(
	config DynamoDBGroupEventRepositoryConfig,
	timeSource clock.TimeSource,
	db *db.Client,
) *DynamoDBGroupEventRepository {
	return &DynamoDBGroupEventRepository{
		config:     config,
		timeSource: timeSource,
		db:         db,
	}
}

func (r DynamoDBGroupEventRepository) PaginatedEvents(ctx context.Context, groupID string, filters PaginatedEventsFilters) ([]models.MeetupEvent, *PaginatedEventsFilters, error) {
	var after time.Time
	if filters.After != nil {
		after = *filters.After
	} else {
		after = r.timeSource.Now().UTC()
	}

	keyCond := expression.Key("groupId").
		Equal(expression.Value(groupID)).
		And(expression.Key("dateTime").
			GreaterThan(expression.Value(after)))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()

	if err != nil {
		return nil, nil, err
	}

	paginator := dynamodb.NewQueryPaginator(r.db, &dynamodb.QueryInput{
		TableName:                 aws.String(r.config.EventsTableName),
		IndexName:                 aws.String(r.config.GroupDateIndexName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	var allEvents []models.MeetupEvent

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, err
		}

		var events []models.MeetupEvent
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &events); err != nil {
			return nil, nil, err
		}
		allEvents = append(allEvents, events...)
	}

	return allEvents, nil, nil
}

var GroupEventRepositoryProviders = wire.NewSet(
	wire.Bind(new(GroupEventRepository), new(*DynamoDBGroupEventRepository)),
	NewDynamoDBGroupEventRepositoryConfig,
	NewDynamoDBGroupEventRepository,
)
