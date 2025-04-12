package importer

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"log"
	"log/slog"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/shared/clock"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"sgf-meetup-api/pkg/syncdynamodb"
	"testing"
	"time"
)

type dbContainer struct {
	db            *dynamodb.Client
	testContainer *tcdynamodb.DynamoDBContainer
}

func (ctr *dbContainer) Close() {
	if ctrErr := testcontainers.TerminateContainer(ctr.testContainer); ctrErr != nil {
		log.Printf("failed to terminate container: %s", ctrErr)
	}
}

func createDb(ctx context.Context) (*dbContainer, error) {
	ctr, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:2.2.1")

	if err != nil {
		return nil, err
	}

	connectionString, err := ctr.ConnectionString(ctx)

	if err != nil {
		return nil, err
	}

	dbOptions := db.Config{
		Endpoint:        "http://" + connectionString,
		Region:          "us-east-2",
		SecretAccessKey: "test",
		AccessKey:       "test",
	}

	db, err := db.New(ctx, dbOptions)

	if err != nil {
		return nil, err
	}

	if err = syncdynamodb.SyncTables(ctx, db, slog.New(logging.NewMockHandler())); err != nil {
		return nil, err
	}

	containerDetails := dbContainer{
		db:            db,
		testContainer: ctr,
	}

	return &containerDetails, nil
}

func TestDynamoDb(t *testing.T) {
	ctx := context.Background()

	dbCtr, err := createDb(ctx)

	if err != nil {
		log.Fatalf("err creating container: %v", err)
	}

	defer dbCtr.Close()

	timeSource := clock.MockTimeSource(time.Now())
	logger := slog.New(logging.NewMockHandler())

	eventsDBRepo := NewEventDBRepository(
		EventDBRepositoryConfig{
			*infra.EventsTableProps.TableName,
			*infra.GroupUrlNameDateTimeIndex.IndexName,
		},
		dbCtr.db,
		timeSource,
		logger,
	)

	service := NewService(
		ServiceConfig{},
		clock.RealTimeSource(),
		logger,
		eventsDBRepo,
		nil,
	)

	err = service.Import(ctx)

	if err != nil {
		t.Fatalf("err creating container: %v", err)
	}
}
