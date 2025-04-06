package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type meetupProxyGraphQLHandler struct {
	proxyFunctionName string
}

func NewMeetupProxyGraphQLHandler(proxyFunctionName string) GraphQLHandler {
	return &meetupProxyGraphQLHandler{
		proxyFunctionName: proxyFunctionName,
	}
}

func (m *meetupProxyGraphQLHandler) ExecuteQuery(ctx context.Context, query string, variables map[string]any) ([]byte, error) {
	request := struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}{
		Query:     query,
		Variables: variables,
	}

	requestBytes, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	return callLambda(ctx, m.proxyFunctionName, requestBytes)
}

func callLambda(ctx context.Context, functionName string, payload []byte) ([]byte, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	result, err := lambda.NewFromConfig(cfg).Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      payload,
	})

	if err != nil {
		return nil, err
	}

	if result.FunctionError != nil {
		return nil, fmt.Errorf("lambda execution error: %s", *result.FunctionError)
	}

	return result.Payload, nil
}
