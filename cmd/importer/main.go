package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"sgf-meetup-api/pkg/importer"
	"time"
)

var service *importer.Service

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newService, err := importer.InitService(ctx)

	if err != nil {
		log.Fatal(err)
	}

	service = newService
}

func main() {
	lambda.Start(service.Import)
}
