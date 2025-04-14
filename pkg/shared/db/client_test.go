package db

import (
	"context"
	"sgf-meetup-api/pkg/shared/logging"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_WithMinimumValidConfig(t *testing.T) {
	client, err := NewClient(context.Background(), Config{
		Region: "us-east-1",
	}, logging.NewMockLogger())

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNew_WithCustomEndpoint(t *testing.T) {
	client, err := NewClient(context.Background(), Config{
		Region:   "us-east-1",
		Endpoint: "http://localhost:8000",
	}, logging.NewMockLogger())

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNew_WithStaticCredentials(t *testing.T) {
	client, err := NewClient(context.Background(), Config{
		Region:          "us-east-1",
		AccessKey:       "AKIAEXAMPLE",
		SecretAccessKey: "SecretExample",
	}, logging.NewMockLogger())

	require.NoError(t, err)
	assert.NotNil(t, client)
}
