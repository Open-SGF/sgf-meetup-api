package db

import (
	"context"
	"testing"

	"sgf-meetup-api/pkg/shared/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_WithMinimumValidConfig(t *testing.T) {
	client, err := NewClient(context.Background(), Config{
		Region:    "us-east-1",
		AccessKey: "access_key",
	}, nil, logging.NewMockLogger())

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNew_WithCustomEndpoint(t *testing.T) {
	client, err := NewClient(context.Background(), Config{
		Region:    "us-east-1",
		AccessKey: "access_key",
		Endpoint:  "http://localhost:8000",
	}, nil, logging.NewMockLogger())

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNew_WithStaticCredentials(t *testing.T) {
	client, err := NewClient(context.Background(), Config{
		Region:          "us-east-1",
		AccessKey:       "AKIAEXAMPLE",
		SecretAccessKey: "SecretExample",
	}, nil, logging.NewMockLogger())

	require.NoError(t, err)
	assert.NotNil(t, client)
}
