package main

import (
	"context"
	"log"
	"time"

	"sgf-meetup-api/pkg/importer"

	"github.com/aws/aws-lambda-go/lambda"
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
