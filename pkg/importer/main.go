package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

func Import(ctx context.Context, config Config) error {
	//token, err := callGetTokenLambda(ctx, config.MeetupProxyFunctionName)

	//if err != nil {
	//	return err
	//}

	//_, err = db.New(ctx, &db.Options{
	//	Endpoint:        config.DynamoDbEndpoint,
	//	Region:          config.AwsRegion,
	//	AccessKey:       config.AwsAccessKey,
	//	SecretAccessKey: config.AwsSecretAccessKey,
	//})

	//if err != nil {
	//	return err
	//}

	//NewMeetupClient(ctx, config.MeetupApiUrl, token)

	//client.GetFutureEvents()

	//log.Println("token", token)

	return nil
}

func callGetTokenLambda(ctx context.Context, functionName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	result, err := lambda.NewFromConfig(cfg).Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
	})

	if err != nil {
		return "", fmt.Errorf("error invoking Lambda function: %w", err)
	}

	if result.FunctionError != nil {
		return "", fmt.Errorf("lambda execution error: %s", *result.FunctionError)
	}

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
