package db

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"log"
	"log/slog"
	"sgf-meetup-api/pkg/shared/logging"
)

type TestDB struct {
	Client    *Client
	Container *tcdynamodb.DynamoDBContainer
}

func (ctr *TestDB) Close() {
	if ctrErr := testcontainers.TerminateContainer(ctr.Container); ctrErr != nil {
		log.Printf("failed to terminate Container: %s", ctrErr)
	}
}

func NewTestDB(ctx context.Context) (*TestDB, error) {
	ctr, err := tcdynamodb.Run(
		ctx,
		"amazon/dynamodb-local:2.6.0",
		tcdynamodb.WithSharedDB(),
		tcdynamodb.WithDisableTelemetry(),
	)

	if err != nil {
		return nil, err
	}

	connectionString, err := ctr.ConnectionString(ctx)

	if err != nil {
		return nil, err
	}

	dbOptions := Config{
		Endpoint:        "http://" + connectionString,
		Region:          "us-east-2",
		SecretAccessKey: "test",
		AccessKey:       "test",
	}

	logger := logging.DefaultLogger(logging.Config{Level: slog.LevelError, Type: logging.LogTypeText})

	client, err := NewClient(ctx, dbOptions, logger)

	if err != nil {
		return nil, err
	}

	if err = SyncTables(ctx, logger, client, Tables); err != nil {
		return nil, err
	}

	testDB := TestDB{
		Client:    client,
		Container: ctr,
	}

	return &testDB, nil
}
