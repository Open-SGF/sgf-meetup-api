package importer

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"sgf-meetup-api/pkg/shared/models"
	"testing"
	"time"
)

func TestEventDBRepository_GetUpcomingEventsForGroup(t *testing.T) {
	ctx := context.Background()
	testDB, err := db.NewTestDB(ctx)
	if err != nil {
		require.NoError(t, err)
	}
	defer testDB.Close()

	mockNow := time.Date(2025, 4, 12, 10, 0, 0, 0, time.UTC)
	timeSource := clock.NewMockTimeSource(mockNow)
	repoConfig := DynamoDBEventRepositoryConfig{
		EventsTableName:    *infra.EventsTableProps.TableName,
		GroupDateIndexName: *infra.GroupIdDateTimeIndex.IndexName,
	}

	repo := NewDynamoDBEventRepository(
		repoConfig,
		testDB.Client,
		timeSource,
		logging.NewMockLogger(),
	)

	faker := gofakeit.New(0)

	t.Run("returns empty slice when no events exist", func(t *testing.T) {
		defer deleteAllItems(t, testDB.Client, repoConfig.EventsTableName)

		events, err := repo.GetUpcomingEventsForGroup(ctx, "test-group")
		require.NoError(t, err)
		require.Empty(t, events)
	})

	t.Run("returns only future events for target group", func(t *testing.T) {
		defer deleteAllItems(t, testDB.Client, repoConfig.EventsTableName)

		testEvents := []models.MeetupEvent{
			createEvent(faker, "test-group", mockNow.Add(-1*time.Hour)),
			createEvent(faker, "test-group", mockNow.Add(1*time.Hour)),
			createEvent(faker, "other-group", mockNow.Add(2*time.Hour)),
			createEvent(faker, "test-group", mockNow.Add(3*time.Hour)),
		}

		insertTestEvents(t, testDB.Client, repoConfig.EventsTableName, testEvents)

		result, err := repo.GetUpcomingEventsForGroup(ctx, "test-group")
		require.NoError(t, err)
		assert.Len(t, result, 2)

		for _, event := range result {
			assert.Equal(t, "test-group", event.GroupId)
			assert.True(t, event.DateTime.After(mockNow))
		}
	})

	t.Run("handles paginated results", func(t *testing.T) {
		defer deleteAllItems(t, testDB.Client, repoConfig.EventsTableName)

		var testEvents []models.MeetupEvent
		for i := 0; i < 15; i++ {
			testEvents = append(testEvents, createEvent(
				faker,
				"test-group",
				mockNow.Add(time.Duration(i+1)*time.Hour),
			))
		}

		insertTestEvents(t, testDB.Client, repoConfig.EventsTableName, testEvents)

		result, err := repo.GetUpcomingEventsForGroup(ctx, "test-group")
		require.NoError(t, err)
		assert.Len(t, result, 15)
	})

	t.Run("excludes events from other groups", func(t *testing.T) {
		defer deleteAllItems(t, testDB.Client, repoConfig.EventsTableName)

		testEvents := []models.MeetupEvent{
			createEvent(faker, "other-group-1", mockNow.Add(1*time.Hour)),
			createEvent(faker, "other-group-2", mockNow.Add(2*time.Hour)),
			createEvent(faker, "test-group", mockNow.Add(3*time.Hour)),
		}

		insertTestEvents(t, testDB.Client, repoConfig.EventsTableName, testEvents)

		result, err := repo.GetUpcomingEventsForGroup(ctx, "test-group")
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestEventDBRepository_ArchiveEvents(t *testing.T) {
	ctx := context.Background()
	testDB, err := db.NewTestDB(ctx)
	require.NoError(t, err)
	defer testDB.Close()

	mockNow := time.Date(2025, 4, 12, 10, 0, 0, 0, time.UTC)
	timeSource := clock.NewMockTimeSource(mockNow)
	repoConfig := DynamoDBEventRepositoryConfig{
		EventsTableName:         *infra.EventsTableProps.TableName,
		ArchivedEventsTableName: *infra.ArchivedEventsTableProps.TableName,
		GroupDateIndexName:      *infra.GroupIdDateTimeIndex.IndexName,
	}

	repo := NewDynamoDBEventRepository(
		repoConfig,
		testDB.Client,
		timeSource,
		logging.NewMockLogger(),
	)

	faker := gofakeit.New(0)

	t.Run("archives and deletes multiple events", func(t *testing.T) {
		defer deleteAllItems(t, testDB.Client, repoConfig.EventsTableName)
		defer deleteAllItems(t, testDB.Client, repoConfig.ArchivedEventsTableName)

		testEvents := []models.MeetupEvent{
			createEvent(faker, "test-group", mockNow.Add(1*time.Hour)),
			createEvent(faker, "test-group", mockNow.Add(2*time.Hour)),
			createEvent(faker, "test-group", mockNow.Add(3*time.Hour)),
		}
		insertTestEvents(t, testDB.Client, repoConfig.EventsTableName, testEvents)

		eventIDs := []string{testEvents[0].ID, testEvents[1].ID, testEvents[2].ID}
		require.NoError(t, repo.ArchiveEvents(ctx, eventIDs))

		for _, id := range eventIDs {
			assert.False(t, checkEventExists(t, testDB.Client, repoConfig.EventsTableName, id), "event should be deleted from main table")
			assert.True(t, checkEventExists(t, testDB.Client, repoConfig.ArchivedEventsTableName, id), "event should exist in archive table")
		}
	})

	t.Run("handles empty input list", func(t *testing.T) {
		defer deleteAllItems(t, testDB.Client, repoConfig.EventsTableName)
		defer deleteAllItems(t, testDB.Client, repoConfig.ArchivedEventsTableName)

		require.NoError(t, repo.ArchiveEvents(ctx, []string{}))
		require.NoError(t, repo.ArchiveEvents(ctx, nil))
	})

	t.Run("handles partial failures gracefully", func(t *testing.T) {
		defer deleteAllItems(t, testDB.Client, repoConfig.EventsTableName)
		defer deleteAllItems(t, testDB.Client, repoConfig.ArchivedEventsTableName)

		validEvent := createEvent(faker, "test-group", mockNow.Add(1*time.Hour))
		insertTestEvents(t, testDB.Client, *infra.EventsTableProps.TableName, []models.MeetupEvent{validEvent})

		err := repo.ArchiveEvents(ctx, []string{validEvent.ID, "non-existent-id"})
		require.NoError(t, err)

		assert.False(t, checkEventExists(t, testDB.Client, repoConfig.EventsTableName, validEvent.ID))
		assert.True(t, checkEventExists(t, testDB.Client, repoConfig.ArchivedEventsTableName, validEvent.ID))
	})

	t.Run("handles large batches with chunking", func(t *testing.T) {
		defer deleteAllItems(t, testDB.Client, repoConfig.EventsTableName)
		defer deleteAllItems(t, testDB.Client, repoConfig.ArchivedEventsTableName)

		var eventIDs []string
		var events []models.MeetupEvent
		for i := 0; i < 30; i++ {
			event := createEvent(faker, "test-group", mockNow.Add(time.Duration(i)*time.Hour))
			events = append(events, event)
			eventIDs = append(eventIDs, event.ID)
		}
		insertTestEvents(t, testDB.Client, repoConfig.EventsTableName, events)

		require.NoError(t, repo.ArchiveEvents(ctx, eventIDs))

		for _, id := range eventIDs {
			assert.False(t, checkEventExists(t, testDB.Client, repoConfig.EventsTableName, id))
			assert.True(t, checkEventExists(t, testDB.Client, repoConfig.ArchivedEventsTableName, id))
		}
	})
}

func createEvent(faker *gofakeit.Faker, groupId string, dateTime time.Time) models.MeetupEvent {
	event := models.MeetupEvent{}
	_ = faker.Struct(&event)
	event.GroupId = groupId
	event.DateTime = &dateTime
	return event
}

func deleteAllItems(t *testing.T, client *db.Client, tableName string) {
	scanInput := &dynamodb.ScanInput{TableName: aws.String(tableName)}
	paginator := dynamodb.NewScanPaginator(client, scanInput)
	ctx := context.Background()

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		require.NoError(t, err)

		for _, item := range page.Items {
			idAttr := item["id"].(*types.AttributeValueMemberS)
			_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
				TableName: aws.String(tableName),
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: idAttr.Value},
				},
			})
			require.NoError(t, err)
		}
	}
}

func insertTestEvents(t *testing.T, client *db.Client, tableName string, events []models.MeetupEvent) {
	for _, event := range events {
		av, err := attributevalue.MarshalMap(event)
		require.NoError(t, err)

		_, err = client.PutItem(context.Background(), &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      av,
		})
		require.NoError(t, err)
	}
}

func checkEventExists(t *testing.T, client *db.Client, tableName string, eventID string) bool {
	resp, err := client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: eventID},
		},
	})
	require.NoError(t, err)
	return len(resp.Item) > 0
}
