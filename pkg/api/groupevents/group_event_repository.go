package groupevents

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/api/apiconfig"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/models"
	"strings"
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
	NextEvent(ctx *gin.Context, groupID string) (*models.MeetupEvent, error)
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

func (r *DynamoDBGroupEventRepository) PaginatedEvents(ctx context.Context, groupID string, filters PaginatedEventsFilters) ([]models.MeetupEvent, *PaginatedEventsFilters, error) {
	keyCond := expression.Key("groupId").
		Equal(expression.Value(groupID))

	switch {
	case filters.After != nil && filters.Before != nil:
		keyCond = keyCond.And(expression.Key("dateTime").Between(
			expression.Value(*filters.After),
			expression.Value(*filters.Before),
		))
	case filters.After != nil:
		keyCond = keyCond.And(expression.Key("dateTime").GreaterThan(expression.Value(*filters.After)))
	case filters.Before != nil:
		keyCond = keyCond.And(expression.Key("dateTime").LessThan(expression.Value(*filters.Before)))
	default:
		now := r.timeSource.Now().UTC()
		keyCond = keyCond.And(expression.Key("dateTime").GreaterThan(expression.Value(now)))
	}

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, nil, err
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.config.EventsTableName),
		IndexName:                 aws.String(r.config.GroupDateIndexName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	}

	if filters.Cursor != "" {
		startKey, err := decodeCursor(filters.Cursor)
		if err != nil {
			return nil, nil, err
		}
		queryInput.ExclusiveStartKey = startKey
	}

	if filters.Limit != nil {
		queryInput.Limit = aws.Int32(int32(*filters.Limit))
	}

	result, err := r.db.Query(ctx, queryInput)
	if err != nil {
		return nil, nil, err
	}

	var events []models.MeetupEvent
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &events); err != nil {
		return nil, nil, err
	}

	var nextCursor *PaginatedEventsFilters
	if result.LastEvaluatedKey != nil {
		cursorStr, err := encodeCursor(result.LastEvaluatedKey)
		if err != nil {
			return nil, nil, err
		}
		nextCursor = &PaginatedEventsFilters{
			Before: filters.Before,
			After:  filters.After,
			Limit:  filters.Limit,
			Cursor: cursorStr,
		}
	}

	return events, nextCursor, nil
}

func (r *DynamoDBGroupEventRepository) NextEvent(ctx *gin.Context, groupID string) (*models.MeetupEvent, error) {
	now := r.timeSource.Now().UTC()

	keyCond := expression.Key("groupId").
		Equal(expression.Value(groupID)).
		And(expression.Key("dateTime").
			GreaterThan(expression.Value(now)))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.config.EventsTableName),
		IndexName:                 aws.String(r.config.GroupDateIndexName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		Limit:                     aws.Int32(1),
	}

	result, err := r.db.Query(ctx, queryInput)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, ErrEventNotFound
	}

	var event models.MeetupEvent
	if err := attributevalue.UnmarshalMap(result.Items[0], &event); err != nil {
		return nil, err
	}

	return &event, nil
}

func encodeCursor(lastKey map[string]types.AttributeValue) (string, error) {
	var id string
	if err := attributevalue.Unmarshal(lastKey["id"], &id); err != nil {
		return "", err
	}

	var groupID string
	if err := attributevalue.Unmarshal(lastKey["groupId"], &groupID); err != nil {
		return "", err
	}

	var dateTime string
	if err := attributevalue.Unmarshal(lastKey["dateTime"], &dateTime); err != nil {
		return "", err
	}

	encodedID := base64.URLEncoding.EncodeToString([]byte(id))
	encodedGroup := base64.URLEncoding.EncodeToString([]byte(groupID))
	encodedTime := base64.URLEncoding.EncodeToString([]byte(dateTime))
	return encodedID + "." + encodedGroup + "." + encodedTime, nil
}

func decodeCursor(cursorStr string) (map[string]types.AttributeValue, error) {
	parts := strings.Split(cursorStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidCursor
	}

	idBytes, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidCursor
	}

	groupIDBytes, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidCursor
	}

	dateTimeBytes, err := base64.URLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidCursor
	}

	return map[string]types.AttributeValue{
		"id":       &types.AttributeValueMemberS{Value: string(idBytes)},
		"groupId":  &types.AttributeValueMemberS{Value: string(groupIDBytes)},
		"dateTime": &types.AttributeValueMemberS{Value: string(dateTimeBytes)},
	}, nil
}

var ErrInvalidCursor = errors.New("invalid cursor")
var ErrEventNotFound = errors.New("event not found")

var GroupEventRepositoryProviders = wire.NewSet(
	wire.Bind(new(GroupEventRepository), new(*DynamoDBGroupEventRepository)),
	NewDynamoDBGroupEventRepositoryConfig,
	NewDynamoDBGroupEventRepository,
)
