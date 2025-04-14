package db

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"net/url"
	"strings"
)

type Config struct {
	Endpoint        string
	Region          string
	AccessKey       string
	SecretAccessKey string
}

func (c Config) ResolveEndpoint(ctx context.Context, params dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
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
