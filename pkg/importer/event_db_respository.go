package importer

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log/slog"
	"sgf-meetup-api/pkg/clock"
	"sgf-meetup-api/pkg/db"
	"sgf-meetup-api/pkg/models"
	"sync"
	"time"
)

type EventDBRepository interface {
	GetUpcomingEventsForGroup(ctx context.Context, group string) ([]models.MeetupEvent, error)
}

type eventDBRepository struct {
	tableName      string
	groupDateIndex string
	dbOptions      db.Options
	mu             sync.Mutex
	db             *dynamodb.Client
	timeSource     clock.TimeSource
	logger         *slog.Logger
}

func NewEventDBRepository(tableName string, groupDateIndex string, dbOptions db.Options, timeSource clock.TimeSource, logger *slog.Logger) EventDBRepository {
	return &eventDBRepository{
		tableName:      tableName,
		groupDateIndex: groupDateIndex,
		dbOptions:      dbOptions,
		timeSource:     timeSource,
		logger:         logger,
	}
}

func (er *eventDBRepository) GetUpcomingEventsForGroup(ctx context.Context, group string) ([]models.MeetupEvent, error) {
	if err := er.initDB(ctx); err != nil {
		return nil, err
	}

	now := er.timeSource.Now().UTC().Format(time.RFC3339)

	var allEvents []models.MeetupEvent

	keyCond := expression.Key("group.name").
		Equal(expression.Value(group)).
		And(expression.Key("dateTime").
			GreaterThan(expression.Value(now)))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()

	if err != nil {
		return nil, err
	}

	paginator := dynamodb.NewQueryPaginator(er.db, &dynamodb.QueryInput{
		TableName:                 aws.String(er.tableName),
		IndexName:                 aws.String(er.groupDateIndex),
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

func (er *eventDBRepository) initDB(ctx context.Context) error {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.db == nil {
		newDb, err := db.New(ctx, &er.dbOptions)

		if err != nil {
			return err
		}

		er.db = newDb
	}

	return nil
}
