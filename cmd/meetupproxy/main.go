package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"sgf-meetup-api/pkg/meetupproxy"
)

var service *meetupproxy.Service

func init() {
	newService, err := meetupproxy.InitService()

	if err != nil {
		log.Fatal(err)
	}

	service = newService
}

func main() {
	lambda.Start(service.HandleRequest)
}
