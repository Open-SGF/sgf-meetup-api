package db

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log/slog"
)

const (
	MaxBatchSize = 25
)

type Client struct {
	*dynamodb.Client
}

func NewClient(ctx context.Context, cfg Config, logger *slog.Logger) (*Client, error) {
	var cfgOpts []func(*config.LoadOptions) error
	var clientOpts []func(*dynamodb.Options)

	if cfg.Region != "" {
		cfgOpts = append(cfgOpts, config.WithRegion(cfg.Region))
	}

	if cfg.AccessKey != "" && cfg.SecretAccessKey != "" {
		cfgOpts = append(cfgOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretAccessKey,
				"",
			),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)

	if err != nil {
		logger.Error("Failed to create dynamo db instance", "err", err)
		return nil, err
	}

	if cfg.Endpoint != "" {
		clientOpts = append(clientOpts, dynamodb.WithEndpointResolverV2(cfg))
	}

	dynamoDbClient := dynamodb.NewFromConfig(awsCfg, clientOpts...)

	return &Client{dynamoDbClient}, nil
}
