package db

import (
	"context"
	"sgf-meetup-api/pkg/shared/logging"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_WithMinimumValidConfig(t *testing.T) {
	client, err := New(context.Background(), Config{
		Region: "us-east-1",
	}, logging.NewMockLogger())

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNew_WithCustomEndpoint(t *testing.T) {
	client, err := New(context.Background(), Config{
		Region:   "us-east-1",
		Endpoint: "http://localhost:8000",
	}, logging.NewMockLogger())

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNew_WithStaticCredentials(t *testing.T) {
	client, err := New(context.Background(), Config{
		Region:          "us-east-1",
		AccessKey:       "AKIAEXAMPLE",
		SecretAccessKey: "SecretExample",
	}, logging.NewMockLogger())

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestResolveEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       string
		expectedScheme string
		expectedHost   string
	}{
		{
			name:           "HTTP endpoint",
			endpoint:       "http://localhost:8000",
			expectedScheme: "http",
			expectedHost:   "localhost:8000",
		},
		{
			name:           "HTTPS endpoint",
			endpoint:       "https://dynamodb.local",
			expectedScheme: "https",
			expectedHost:   "dynamodb.local",
		},
		{
			name:           "No scheme",
			endpoint:       "localhost:8000",
			expectedScheme: "",
			expectedHost:   "localhost:8000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{Endpoint: tt.endpoint}
			ep, err := c.ResolveEndpoint(context.Background(), dynamodb.EndpointParameters{})

			require.NoError(t, err)
			assert.Equal(t, tt.expectedScheme, ep.URI.Scheme)
			assert.Equal(t, tt.expectedHost, ep.URI.Host)
		})
	}
}
