package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"sgf-meetup-api/pkg/importer"
)

var config *importer.Config
var service *importer.Service

func init() {
	cfg, err := importer.NewConfig()

	if err != nil {
		log.Fatal(err)
	}

	config = cfg
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context) error {
	if service == nil {
		newService, err := importer.InitService(ctx, config)

		if err != nil {
			return err
		}

		service = newService
	}

	return service.Import(ctx)
}
