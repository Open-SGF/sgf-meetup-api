package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"sgf-meetup-api/src/importer"
)

var config *importer.Config

func init() {
	config = importer.LoadConfig()
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, event json.RawMessage) error {
	err := importer.Import(ctx, *config)

	return err
}
