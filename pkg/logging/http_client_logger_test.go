package logging

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHttpLoggingTransport_SuccessfulRequest(t *testing.T) {
	mockHandler := NewMockHandler()
	transport := NewHttpLoggingTransport(slog.New(mockHandler))
	client := &http.Client{Transport: transport}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1234")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Unexpected error making request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	infoEntries := mockHandler.Entries(slog.LevelInfo)
	if len(infoEntries) != 1 {
		t.Fatalf("Expected 1 info log entry, got %d", len(infoEntries))
	}

	entry := infoEntries[0]
	if entry.Message != "request completed" {
		t.Errorf("Unexpected message: %q", entry.Message)
	}

	expectedAttrs := map[string]any{
		"http_client.method":         "GET",
		"http_client.url":            ts.URL,
		"http_client.status_code":    int64(200),
		"http_client.content_length": "1234",
	}

	for k, v := range expectedAttrs {
		if entry.Attrs[k] != v {
			t.Errorf("Attribute mismatch for %q: expected %v, got %v", k, v, entry.Attrs[k])
		}
	}

	if _, ok := entry.Attrs["http_client.duration"].(time.Duration); !ok {
		t.Error("Missing or invalid duration attribute")
	}
}

func TestHttpLoggingTransport_FailedRequest(t *testing.T) {
	mockHandler := NewMockHandler()
	transport := NewHttpLoggingTransport(slog.New(mockHandler))
	client := &http.Client{Transport: transport}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	_, err := client.Do(req)
	if err == nil {
		t.Fatal("Expected error from closed server, got nil")
	}

	errorEntries := mockHandler.Entries(slog.LevelError)
	if len(errorEntries) != 1 {
		t.Fatalf("Expected 1 error log entry, got %d", len(errorEntries))
	}

	entry := errorEntries[0]
	if entry.Message != "request failed" {
		t.Errorf("Unexpected message: %q", entry.Message)
	}

	if _, ok := entry.Attrs["http_client.method"].(string); !ok {
		t.Error("Missing method in error log")
	}

	if _, ok := entry.Attrs["http_client.url"].(string); !ok {
		t.Error("Missing URL in error log")
	}

	if _, ok := entry.Attrs["http_client.error"].(error); !ok {
		t.Error("Missing error in error log")
	}
}
