package importer

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sgf-meetup-api/pkg/logging"
	"testing"
)

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

	handler := NewMeetupProxyGraphQLHandler("test-function", slog.New(logging.NewMockHandler()))

	result, err := handler.ExecuteQuery(context.Background(), "query {}", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(result, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
}

func TestExecuteQuery_LambdaExecutionError(t *testing.T) {
	testServer := setupMockLambdaServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Amz-Function-Error", "Unhandled")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "lambda failure"})
	})
	defer testServer.Close()

	handler := NewMeetupProxyGraphQLHandler("test-function", slog.New(logging.NewMockHandler()))

	_, err := handler.ExecuteQuery(context.Background(), "query {}", nil)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	expectedErr := "lambda execution error: Unhandled"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s' but got '%s'", expectedErr, err.Error())
	}
}

func TestExecuteQuery_LambdaInvokeError(t *testing.T) {
	testServer := setupMockLambdaServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer testServer.Close()

	handler := NewMeetupProxyGraphQLHandler("test-function", slog.New(logging.NewMockHandler()))

	_, err := handler.ExecuteQuery(context.Background(), "query {}", nil)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
}

func TestExecuteQuery_MarshalError(t *testing.T) {
	handler := NewMeetupProxyGraphQLHandler("test-function", slog.New(logging.NewMockHandler()))

	invalidVariables := map[string]any{
		"channel": make(chan int), // Channels can't be JSON marshaled
	}

	_, err := handler.ExecuteQuery(context.Background(), "query {}", invalidVariables)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
}
