package importer

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sgf-meetup-api/pkg/importer/importerconfig"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/fakers"
	"sgf-meetup-api/pkg/shared/logging"
	"sgf-meetup-api/pkg/shared/models"
	"testing"
	"time"
)

func TestNewDynamoDBEventRepositoryConfig(t *testing.T) {
	cfg := &importerconfig.Config{
		EventsTableName:          "events",
		ArchivedEventsTableName:  "archivedEvents",
		GroupIDDateTimeIndexName: "groupIdDateTimeIndex",
	}

	eventRepoConfig := NewDynamoDBEventRepositoryConfig(cfg)

	assert.Equal(t, cfg.EventsTableName, eventRepoConfig.EventsTableName)
	assert.Equal(t, cfg.ArchivedEventsTableName, eventRepoConfig.ArchivedEventsTableName)
	assert.Equal(t, cfg.GroupIDDateTimeIndexName, eventRepoConfig.GroupDateIndexName)
}

func TestDynamoDBEventRepository_GetUpcomingEventsForGroup(t *testing.T) {
	ctx := context.Background()
	testDB, err := db.NewTestDB(ctx)
	require.NoError(t, err)
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

	meetupFaker := fakers.NewMeetupFaker(0)

	t.Run("returns empty slice when no events exist", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		events, err := repo.GetUpcomingEventsForGroup(ctx, "test-group")
		require.NoError(t, err)
		require.Empty(t, events)
	})

	t.Run("returns only future events for target group", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		testEvents := []models.MeetupEvent{
			meetupFaker.CreateEvent("test-group", mockNow.Add(-1*time.Hour)),
			meetupFaker.CreateEvent("test-group", mockNow.Add(1*time.Hour)),
			meetupFaker.CreateEvent("other-group", mockNow.Add(2*time.Hour)),
			meetupFaker.CreateEvent("test-group", mockNow.Add(3*time.Hour)),
		}

		testDB.InsertTestItems(ctx, repoConfig.EventsTableName, testEvents)

		result, err := repo.GetUpcomingEventsForGroup(ctx, "test-group")
		require.NoError(t, err)
		assert.Len(t, result, 2)

		for _, event := range result {
			assert.Equal(t, "test-group", event.GroupID)
			assert.True(t, event.DateTime.After(mockNow))
		}
	})

	t.Run("handles paginated results", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		var testEvents []models.MeetupEvent
		for i := 0; i < 15; i++ {
			testEvents = append(testEvents, meetupFaker.CreateEvent(
				"test-group",
				mockNow.Add(time.Duration(i+1)*time.Hour),
			))
		}

		testDB.InsertTestItems(ctx, repoConfig.EventsTableName, testEvents)

		result, err := repo.GetUpcomingEventsForGroup(ctx, "test-group")
		require.NoError(t, err)
		assert.Len(t, result, 15)
	})

	t.Run("excludes events from other groups", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		testEvents := []models.MeetupEvent{
			meetupFaker.CreateEvent("other-group-1", mockNow.Add(1*time.Hour)),
			meetupFaker.CreateEvent("other-group-2", mockNow.Add(2*time.Hour)),
			meetupFaker.CreateEvent("test-group", mockNow.Add(3*time.Hour)),
		}

		testDB.InsertTestItems(ctx, repoConfig.EventsTableName, testEvents)

		result, err := repo.GetUpcomingEventsForGroup(ctx, "test-group")
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestDynamoDBEventRepository_ArchiveEvents(t *testing.T) {
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

	meetupFaker := fakers.NewMeetupFaker(0)

	t.Run("archives and deletes multiple events", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		testEvents := []models.MeetupEvent{
			meetupFaker.CreateEvent("test-group", mockNow.Add(1*time.Hour)),
			meetupFaker.CreateEvent("test-group", mockNow.Add(2*time.Hour)),
			meetupFaker.CreateEvent("test-group", mockNow.Add(3*time.Hour)),
		}
		testDB.InsertTestItems(ctx, repoConfig.EventsTableName, testEvents)

		eventIDs := []string{testEvents[0].ID, testEvents[1].ID, testEvents[2].ID}
		require.NoError(t, repo.ArchiveEvents(ctx, eventIDs))

		for _, id := range eventIDs {
			assert.False(t, testDB.CheckItemExists(ctx, repoConfig.EventsTableName, "id", id), "event should be deleted from main table")
			assert.True(t, testDB.CheckItemExists(ctx, repoConfig.ArchivedEventsTableName, "id", id), "event should exist in archive table")
		}
	})

	t.Run("handles empty input list", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		require.NoError(t, repo.ArchiveEvents(ctx, []string{}))
		require.NoError(t, repo.ArchiveEvents(ctx, nil))
	})

	t.Run("handles partial failures gracefully", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		validEvent := meetupFaker.CreateEvent("test-group", mockNow.Add(1*time.Hour))
		testDB.InsertTestItems(ctx, repoConfig.EventsTableName, []models.MeetupEvent{validEvent})

		err = repo.ArchiveEvents(ctx, []string{validEvent.ID, "non-existent-id"})
		require.NoError(t, err)

		assert.False(t, testDB.CheckItemExists(ctx, repoConfig.EventsTableName, "id", validEvent.ID))
		assert.True(t, testDB.CheckItemExists(ctx, repoConfig.ArchivedEventsTableName, "id", validEvent.ID))
	})

	t.Run("handles large batches with chunking", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		var eventIDs []string
		var events []models.MeetupEvent
		for i := 0; i < db.MaxBatchSize+5; i++ {
			event := meetupFaker.CreateEvent("test-group", mockNow.Add(time.Duration(i)*time.Hour))
			events = append(events, event)
			eventIDs = append(eventIDs, event.ID)
		}
		testDB.InsertTestItems(ctx, repoConfig.EventsTableName, events)

		require.NoError(t, repo.ArchiveEvents(ctx, eventIDs))

		for _, id := range eventIDs {
			assert.False(t, testDB.CheckItemExists(ctx, repoConfig.EventsTableName, "id", id))
			assert.True(t, testDB.CheckItemExists(ctx, repoConfig.ArchivedEventsTableName, "id", id))
		}
	})
}

func TestDynamoDBEventRepository_UpsertEvents(t *testing.T) {
	ctx := context.Background()
	testDB, err := db.NewTestDB(ctx)
	require.NoError(t, err)
	defer testDB.Close()

	repoConfig := DynamoDBEventRepositoryConfig{
		EventsTableName: *infra.EventsTableProps.TableName,
	}
	repo := NewDynamoDBEventRepository(
		repoConfig,
		testDB.Client,
		clock.NewMockTimeSource(time.Now()),
		logging.NewMockLogger(),
	)
	meetupFaker := fakers.NewMeetupFaker(0)

	t.Run("inserts new events into table", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		testEvents := []models.MeetupEvent{
			meetupFaker.CreateEvent("group1", time.Now().Add(1*time.Hour)),
			meetupFaker.CreateEvent("group1", time.Now().Add(2*time.Hour)),
		}

		require.NoError(t, repo.UpsertEvents(ctx, testEvents))

		for _, event := range testEvents {
			assert.True(t, testDB.CheckItemExists(ctx, repoConfig.EventsTableName, "id", event.ID))
		}
	})

	t.Run("updates existing events", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		originalEvent := meetupFaker.CreateEvent("group1", time.Now().Add(1*time.Hour))
		require.NoError(t, repo.UpsertEvents(ctx, []models.MeetupEvent{originalEvent}))

		updatedEvent := originalEvent
		updatedEvent.Title = "UPDATED TITLE"
		require.NoError(t, repo.UpsertEvents(ctx, []models.MeetupEvent{updatedEvent}))

		var result models.MeetupEvent
		resp, err := testDB.Client.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(repoConfig.EventsTableName),
			Key:       repo.createKey(originalEvent.ID),
		})
		require.NoError(t, err)
		require.NoError(t, attributevalue.UnmarshalMap(resp.Item, &result))

		assert.Equal(t, "UPDATED TITLE", result.Title)
	})

	t.Run("handles empty input list", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		require.NoError(t, repo.UpsertEvents(ctx, []models.MeetupEvent{}))
		require.NoError(t, repo.UpsertEvents(ctx, nil))
	})

	t.Run("chunks large batches", func(t *testing.T) {
		defer func() { _ = testDB.Reset(ctx) }()

		var testEvents []models.MeetupEvent
		for i := 0; i < db.MaxBatchSize+5; i++ {
			testEvents = append(testEvents, meetupFaker.CreateEvent("group1", time.Now()))
		}

		require.NoError(t, repo.UpsertEvents(ctx, testEvents))

		eventCount := testDB.GetItemCount(ctx, repoConfig.EventsTableName)
		assert.Equal(t, db.MaxBatchSize+5, eventCount)
	})
}
