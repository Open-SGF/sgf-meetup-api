package db

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Config struct {
	Endpoint        string
	Region          string
	AccessKey       string
	SecretAccessKey string
}

func New(ctx context.Context, options Config) (*dynamodb.Client, error) {
	var cfgOpts []func(*config.LoadOptions) error
	var clientOpts []func(*dynamodb.Options)

	if options.Region != "" {
		cfgOpts = append(cfgOpts, config.WithRegion(options.Region))
	}

	if options.AccessKey != "" && options.SecretAccessKey != "" {
		cfgOpts = append(cfgOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				options.AccessKey,
				options.SecretAccessKey,
				"",
			),
		))
	}

	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)

	if err != nil {
		return nil, err
	}

	if options.Endpoint != "" {
		clientOpts = append(clientOpts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(options.Endpoint)
			o.EndpointOptions.DisableHTTPS = true
		})
	}

	return dynamodb.NewFromConfig(cfg, clientOpts...), nil
}
