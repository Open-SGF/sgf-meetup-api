package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"sgf-meetup-api/src/meetuptoken"
)

var config *meetuptoken.Config

func init() {
	config = meetuptoken.LoadConfig()
}

func main() {
	lambda.Start(handleRequest)
}

type Request struct {
	ClientId string `json:"clientId"`
}

type Response struct {
	Token        string `json:"token,omitempty"`
	ErrorName    string `json:"errorName,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func handleRequest(ctx context.Context, request Request) (Response, error) {
	if request.ClientId == "" {
		return Response{
			ErrorName:    "InvalidParameterException",
			ErrorMessage: "clientId is required",
		}, nil
	}

	token, err := meetuptoken.GetToken(ctx, config, request.ClientId)

	if err != nil {
		return Response{
			ErrorName:    "TokenGenerationException",
			ErrorMessage: fmt.Sprintf("Failed to generate token: %v", err),
		}, nil
	}

	return Response{
		Token: token,
	}, nil
}
