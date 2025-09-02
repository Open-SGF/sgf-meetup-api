package main

import (
	"context"
	"log"
	"time"

	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/syncdynamodb"
)

var service *syncdynamodb.Service

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newService, err := syncdynamodb.InitService(ctx)
	if err != nil {
		log.Fatalf("Failed to init syncdynamdb service: %v", err)
	}

	service = newService
}

func main() {
	if err := service.Run(context.Background(), infra.Tables); err != nil {
		log.Fatalf("Failed to sync database: %v", err)
	}
}
