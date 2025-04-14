package importer

import (
	"context"
	"log"
	"log/slog"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"testing"
	"time"
)

func TestDynamoDb(t *testing.T) {
	ctx := context.Background()

	testDB, err := db.NewTestDB(ctx)

	if err != nil {
		log.Fatalf("err creating container: %v", err)
	}

	defer testDB.Close()

	timeSource := clock.NewMockTimeSource(time.Now())
	logger := slog.New(logging.NewMockHandler())

	eventsDBRepo := NewDynamoDBEventRepository(
		DynamoDBEventRepositoryConfig{
			*infra.EventsTableProps.TableName,
			*infra.ArchivedEventsTableProps.TableName,
			*infra.GroupIdDateTimeIndex.IndexName,
		},
		testDB.Client,
		timeSource,
		logger,
	)

	service := NewService(
		ServiceConfig{},
		clock.NewRealTimeSource(),
		logger,
		eventsDBRepo,
		nil,
	)

	err = service.Import(ctx)

	if err != nil {
		t.Fatalf("err creating container: %v", err)
	}
}
