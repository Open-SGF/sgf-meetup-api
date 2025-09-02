package main

import (
	"context"
	"log"
	"time"

	"sgf-meetup-api/pkg/meetupproxy"

	"github.com/aws/aws-lambda-go/lambda"
)

var service *meetupproxy.Service

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	newService, err := meetupproxy.InitService(ctx)
	if err != nil {
		log.Fatal(err)
	}

	service = newService
}

func main() {
	lambda.Start(service.HandleRequest)
}
