package db

import (
	"context"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

type Config struct {
	Endpoint        string `mapstructure:"dynamodb_endpoint"`
	Region          string `mapstructure:"dynamodb_aws_region"`
	AccessKey       string `mapstructure:"dynamodb_aws_access_key"`
	SecretAccessKey string `mapstructure:"dynamodb_aws_secret_access_key"`
}

func (c Config) ResolveEndpoint(
	ctx context.Context,
	params dynamodb.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	scheme, rest := c.splitEndpoint()

	return smithyendpoints.Endpoint{
		URI: url.URL{Host: rest, Scheme: scheme},
	}, nil
}

func (c Config) splitEndpoint() (scheme, rest string) {
	parts := strings.SplitN(c.Endpoint, "://", 2)
	if len(parts) < 2 {
		return "", c.Endpoint
	}
	return parts[0], parts[1]
}
