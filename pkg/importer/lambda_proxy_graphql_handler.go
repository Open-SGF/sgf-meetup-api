package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"log/slog"
	"sgf-meetup-api/pkg/importer/importerconfig"
)

type LambdaProxyGraphQLHandlerConfig struct {
	ProxyFunctionName string
}

func NewLambdaProxyGraphQLHandlerConfig(config *importerconfig.Config) LambdaProxyGraphQLHandlerConfig {
	return LambdaProxyGraphQLHandlerConfig{
		ProxyFunctionName: config.ProxyFunctionName,
	}
}

type LambdaProxyGraphQLHandler struct {
	config LambdaProxyGraphQLHandlerConfig
	logger *slog.Logger
}

func NewLambdaProxyGraphQLHandler(config LambdaProxyGraphQLHandlerConfig, logger *slog.Logger) *LambdaProxyGraphQLHandler {
	return &LambdaProxyGraphQLHandler{
		config: config,
		logger: logger,
	}
}

func (m *LambdaProxyGraphQLHandler) ExecuteQuery(ctx context.Context, query string, variables map[string]any) ([]byte, error) {
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

func (m *LambdaProxyGraphQLHandler) callLambda(ctx context.Context, payload []byte) ([]byte, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	result, err := lambda.NewFromConfig(cfg).Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(m.config.ProxyFunctionName),
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
