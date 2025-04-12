package db

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"log"
	"sgf-meetup-api/pkg/shared/logging"
)

type TestDB struct {
	DB        *dynamodb.Client
	Container *tcdynamodb.DynamoDBContainer
}

func (ctr *TestDB) Close() {
	if ctrErr := testcontainers.TerminateContainer(ctr.Container); ctrErr != nil {
		log.Printf("failed to terminate Container: %s", ctrErr)
	}
}

func NewTestDB(ctx context.Context) (*TestDB, error) {
	ctr, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:2.2.1")

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

	db, err := New(ctx, dbOptions)

	if err != nil {
		return nil, err
	}

	if err = SyncTables(ctx, db, logging.NewMockLogger()); err != nil {
		return nil, err
	}

	testDB := TestDB{
		DB:        db,
		Container: ctr,
	}

	return &testDB, nil
}
