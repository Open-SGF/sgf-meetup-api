package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"sgf-meetup-api/pkg/importer"
)

var config *importer.Config

func init() {
	cfg, err := importer.NewConfig()

	if err != nil {
		log.Fatal(err)
	}

	config = cfg
}

func main() {
	service := importer.NewFromConfig(config)
	lambda.Start(service.Import)
}
