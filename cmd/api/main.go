package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"sgf-meetup-api/pkg/api"
)

var ginLambda *ginadapter.GinLambda

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	router, err := api.InitRouter(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ginLambda = ginadapter.New(router)
}

func main() {
	lambda.Start(ginLambda.ProxyWithContext)
}
