package main

import (
	"context"
	"log"
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

	service, err := syncdynamodb.InitService(ctx, config)

	if err != nil {
		log.Fatalf("Failed to init syncdynamdb service: %v", err)
	}

	if err := service.Run(ctx); err != nil {
		log.Fatalf("Failed to sync database: %v", err)
	}
}
