package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/wire"
)

const (
	MaxBatchSize = 25
)

type Client struct {
	*dynamodb.Client
}

func NewClient(
	ctx context.Context,
	cfg Config,
	awsCfg *aws.Config,
	logger *slog.Logger,
) (*Client, error) {
	var clientOpts []func(*dynamodb.Options)

	awsCfg, err := getAwsConfig(ctx, cfg, awsCfg)
	if err != nil {
		logger.Error("Failed to create dynamo db instance", "err", err)
		return nil, err
	}

	if cfg.Endpoint != "" {
		clientOpts = append(clientOpts, dynamodb.WithEndpointResolverV2(cfg))
	}

	dynamoDbClient := dynamodb.NewFromConfig(*awsCfg, clientOpts...)

	return &Client{dynamoDbClient}, nil
}

func getAwsConfig(ctx context.Context, cfg Config, awsCfg *aws.Config) (*aws.Config, error) {
	if cfg.AccessKey == "" {
		if awsCfg == nil {
			return nil, fmt.Errorf("no credentials or existing aws config present")
		}

		return awsCfg, nil
	}

	var cfgOpts []func(*config.LoadOptions) error

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

	newAwsCfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return nil, err
	}

	return &newAwsCfg, nil
}

var Providers = wire.NewSet(NewClient)
