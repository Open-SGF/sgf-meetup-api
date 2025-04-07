package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"log/slog"
)

type meetupProxyGraphQLHandler struct {
	proxyFunctionName string
	logger            *slog.Logger
}

func NewMeetupProxyGraphQLHandler(proxyFunctionName string, logger *slog.Logger) GraphQLHandler {
	return &meetupProxyGraphQLHandler{
		proxyFunctionName: proxyFunctionName,
		logger:            logger,
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

	return m.callLambda(ctx, requestBytes)
}

func (m *meetupProxyGraphQLHandler) callLambda(ctx context.Context, payload []byte) ([]byte, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	result, err := lambda.NewFromConfig(cfg).Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(m.proxyFunctionName),
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
