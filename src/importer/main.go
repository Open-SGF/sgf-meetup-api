package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"log"
	"os"
)

const functionNameEnvVar = "GET_MEETUP_TOKEN_FUNCTION_NAME"

func Import(ctx context.Context, config Config) error {
	token, err := getToken(ctx)

	if err != nil {
		return err
	}

	log.Println(token)

	return nil
}

func getToken(ctx context.Context) (string, error) {
	lambdaName := os.Getenv(functionNameEnvVar)
	if lambdaName == "" {
		return "", fmt.Errorf("%s environment variable not set", functionNameEnvVar)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	lambdaClient := lambda.NewFromConfig(cfg)

	result, err := lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName: aws.String(lambdaName),
	})

	if err != nil {
		return "", fmt.Errorf("error invoking Lambda function: %w", err)
	}

	if result.FunctionError != nil {
		return "", fmt.Errorf("Lambda execution error: %s", *result.FunctionError)
	}

	// Parse response
	var response struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(result.Payload, &response); err != nil {
		return "", fmt.Errorf("error parsing Lambda response: %w", err)
	}

	if response.Token == "" {
		return "", fmt.Errorf("token not found in Lambda response")
	}

	return response.Token, nil
}
