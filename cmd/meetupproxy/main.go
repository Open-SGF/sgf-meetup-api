package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"sgf-meetup-api/pkg/meetupproxy"
)

var config *meetupproxy.Config

func init() {
	config = meetupproxy.NewConfig()
}

func main() {
	p := meetupproxy.NewFromConfig(config)
	lambda.Start(p.HandleRequest)
}
