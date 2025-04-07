package importer

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"log"
	"log/slog"
	"sgf-meetup-api/pkg/clock"
	"sgf-meetup-api/pkg/db"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/logging"
	"sgf-meetup-api/pkg/syncdynamodb"
	"testing"
)

type dbContainer struct {
	db.Options
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

	dbOptions := db.Options{
		Endpoint:        "http://" + connectionString,
		Region:          "us-east-2",
		SecretAccessKey: "test",
		AccessKey:       "test",
	}

	db, err := db.New(ctx, &dbOptions)

	if err != nil {
		return nil, err
	}

	if err = syncdynamodb.SyncTables(ctx, db); err != nil {
		return nil, err
	}

	containerDetails := dbContainer{
		Options:       dbOptions,
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

	service := New(
		*infra.EventsTableProps.TableName,
		[]string{},
		dbCtr.Options,
		clock.RealTimeSource(),
		slog.New(logging.NewMockHandler()),
		nil,
	)

	err = service.Import(ctx)

	if err != nil {
		log.Fatalf("err creating container: %v", err)
	}

	log.Println(dbCtr.Endpoint)
}
