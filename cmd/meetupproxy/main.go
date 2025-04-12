package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"sgf-meetup-api/pkg/meetupproxy"
)

var config *meetupproxy.Config

func init() {
	cfg, err := meetupproxy.NewConfig()

	if err != nil {
		log.Fatal(err)
	}

	config = cfg
}

func main() {
	p := meetupproxy.InitService(config)
	lambda.Start(p.HandleRequest)
}
