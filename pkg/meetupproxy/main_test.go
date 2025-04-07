package meetupproxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sgf-meetup-api/pkg/logging"
	"strings"
	"testing"
	"time"
)

type mockAuth struct {
	token string
	err   error
}

func (m *mockAuth) GetAccessToken(ctx context.Context) (string, error) {
	return m.token, m.err
}

func TestProxy_HandleRequest_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": "success"})
	}))

	defer ts.Close()

	proxy := New(ts.URL, slog.New(logging.NewMockHandler()), &mockAuth{token: "valid-token"})

	resp, err := proxy.HandleRequest(context.Background(), Request{Query: "query() {}"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if (*resp)["data"] != "success" {
		t.Errorf("Expected 'success' data, got %v", resp)
	}
}

func TestProxy_HandleRequest_AuthFailure(t *testing.T) {
	auth := &mockAuth{err: fmt.Errorf("auth error")}
	proxy := New("https://testurl", slog.New(logging.NewMockHandler()), auth)

	_, err := proxy.HandleRequest(context.Background(), Request{})

	if err == nil {
		t.Fatalf("expected err but got nil")
	}

	if !strings.Contains(err.Error(), "auth error") {
		t.Fatalf("Expected auth err, got %v", err)
	}
}

func TestProxy_HandleRequest_InvalidJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{invalid json}"))
	}))

	defer ts.Close()

	proxy := New(ts.URL, slog.New(logging.NewMockHandler()), &mockAuth{token: "valid"})

	_, err := proxy.HandleRequest(context.Background(), Request{Query: "query() {}"})

	if err == nil {
		t.Fatalf("expected err but got nil")
	}

	if !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("Expected JSON error, got %v", err)
	}
}

func TestProxy_HandleRequest_InvalidStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "401", statusCode: http.StatusUnauthorized},
		{name: "403", statusCode: http.StatusForbidden},
		{name: "404", statusCode: http.StatusNotFound},
		{name: "500", statusCode: http.StatusInternalServerError},
		{name: "503", statusCode: http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := logging.NewMockHandler()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))

			defer ts.Close()

			proxy := New(ts.URL, slog.New(handler), &mockAuth{token: "valid"})

			_, err := proxy.HandleRequest(context.Background(), Request{Query: "query() {}"})

			if err == nil {
				t.Fatalf("expected err but got nil")
			}

			if !strings.Contains(err.Error(), fmt.Sprintf("expected status code 200, got %v", tt.statusCode)) {
				t.Errorf("Expected status code error, got %v", err)
			}

			errorEntries := handler.Entries(slog.LevelError)
			if len(errorEntries) != 1 {
				t.Errorf("Expected 1 error log entry got %v", len(errorEntries))
			}
		})
	}
}

func TestHandleRequest_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	proxy := New(ts.URL, slog.New(logging.NewMockHandler()), &mockAuth{token: "valid-token"})

	_, err := proxy.HandleRequest(ctx, Request{
		Query: "query() {}",
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected context deadline exceeded, got: %v", err)
	}
}
