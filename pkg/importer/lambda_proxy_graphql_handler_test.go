package importer

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sgf-meetup-api/pkg/importer/importerconfig"
	"sgf-meetup-api/pkg/shared/logging"
	"testing"
)

func TestNewLambdaProxyGraphQLHandlerConfig(t *testing.T) {
	cfg := &importerconfig.Config{
		ProxyFunctionName: "meetupProxy",
	}

	handlerConfig := NewLambdaProxyGraphQLHandlerConfig(cfg)

	assert.Equal(t, cfg.ProxyFunctionName, handlerConfig.ProxyFunctionName)
}

func setupMockLambdaServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2015-03-31/functions/test-function/invocations" {
			t.Errorf("Unexpected API path: %s", r.URL.Path)
		}

		handler(w, r)
	}))

	t.Setenv("AWS_ENDPOINT_URL_LAMBDA", testServer.URL)

	return testServer
}

func TestExecuteQuery_Success(t *testing.T) {
	testServer := setupMockLambdaServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": "test"})
	})
	defer testServer.Close()

	handler := NewLambdaProxyGraphQLHandler(LambdaProxyGraphQLHandlerConfig{"test-function"}, logging.NewMockLogger())

	result, err := handler.ExecuteQuery(context.Background(), "query {}", nil)

	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(result, &response)

	require.NoError(t, err)
}

func TestExecuteQuery_LambdaExecutionError(t *testing.T) {
	testServer := setupMockLambdaServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Amz-Function-Error", "Unhandled")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "lambda failure"})
	})
	defer testServer.Close()

	handler := NewLambdaProxyGraphQLHandler(LambdaProxyGraphQLHandlerConfig{"test-function"}, logging.NewMockLogger())

	_, err := handler.ExecuteQuery(context.Background(), "query {}", nil)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	assert.Equal(t, "lambda execution error: Unhandled", err.Error())
}

func TestExecuteQuery_LambdaInvokeError(t *testing.T) {
	testServer := setupMockLambdaServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer testServer.Close()

	handler := NewLambdaProxyGraphQLHandler(LambdaProxyGraphQLHandlerConfig{"test-function"}, logging.NewMockLogger())

	_, err := handler.ExecuteQuery(context.Background(), "query {}", nil)

	assert.Error(t, err)
}

func TestExecuteQuery_MarshalError(t *testing.T) {
	handler := NewLambdaProxyGraphQLHandler(LambdaProxyGraphQLHandlerConfig{"test-function"}, logging.NewMockLogger())

	invalidVariables := map[string]any{
		"channel": make(chan int), // Channels can't be JSON marshaled
	}

	_, err := handler.ExecuteQuery(context.Background(), "query {}", invalidVariables)

	assert.Error(t, err)
}
