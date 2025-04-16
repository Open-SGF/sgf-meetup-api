package main

import (
	"context"
	"log"
	"sgf-meetup-api/pkg/syncdynamodb"
	"time"
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
	if err := service.Run(context.Background()); err != nil {
		log.Fatalf("Failed to sync database: %v", err)
	}
}
