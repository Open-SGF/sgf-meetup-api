package main

import (
	"context"
	"log"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/logging"
	"sgf-meetup-api/pkg/syncdynamodb"
)

var config *syncdynamodb.Config

func init() {
	cfg, err := syncdynamodb.NewConfig()

	if err != nil {
		log.Fatal(err)
	}

	config = cfg
}

func main() {
	ctx := context.Background()
	client, err := db.New(ctx, db.Config{
		Endpoint:        config.DynamoDbEndpoint,
		Region:          config.AwsRegion,
		AccessKey:       config.AwsAccessKey,
		SecretAccessKey: config.AwsSecretAccessKey,
	})

	if err != nil {
		log.Fatalf("Failed to create DynamoDB client: %v", err)
	}

	if err := syncdynamodb.SyncTables(ctx, client, logging.DefaultLogger(config.LogLevel)); err != nil {
		log.Fatalf("Failed to sync database: %v", err)
	}
}
