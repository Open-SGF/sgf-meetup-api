package db

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"log/slog"
	"net/url"
	"strings"
)

type Config struct {
	Endpoint        string
	Region          string
	AccessKey       string
	SecretAccessKey string
}

func New(ctx context.Context, cfg Config, logger *slog.Logger) (*dynamodb.Client, error) {
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

	return dynamodb.NewFromConfig(awsCfg, clientOpts...), nil
}

func (c Config) ResolveEndpoint(ctx context.Context, params dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
	scheme, rest := splitURL(c.Endpoint)

	return smithyendpoints.Endpoint{
		URI: url.URL{Host: rest, Scheme: scheme},
	}, nil
}

func splitURL(url string) (scheme, rest string) {
	parts := strings.SplitN(url, "://", 2)
	if len(parts) < 2 {
		return "", url
	}
	return parts[0], parts[1]
}
