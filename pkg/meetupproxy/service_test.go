package meetupproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sgf-meetup-api/pkg/meetupproxy/meetupproxyconfig"
	"sgf-meetup-api/pkg/shared/logging"
)

func TestNewServiceConfig(t *testing.T) {
	cfg := &meetupproxyconfig.Config{
		MeetupAPIURL: "https://example.com",
	}

	serviceConfig := NewServiceConfig(cfg)

	assert.Equal(t, cfg.MeetupAPIURL, serviceConfig.URL)
}

type mockAuth struct {
	token string
	err   error
}

func (m *mockAuth) GetAccessToken(ctx context.Context) (string, error) {
	return m.token, m.err
}

func TestService_HandleRequest_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": "success"})
	}))

	defer ts.Close()

	proxy := NewService(
		ServiceConfig{ts.URL},
		&http.Client{},
		&mockAuth{token: "valid-token"},
		logging.NewMockLogger(),
	)

	resp, err := proxy.HandleRequest(context.Background(), Request{Query: "query() {}"})

	require.NoError(t, err)

	assert.Equal(t, "success", (*resp)["data"])
}

func TestService_HandleRequest_AuthFailure(t *testing.T) {
	auth := &mockAuth{err: fmt.Errorf("auth error")}
	proxy := NewService(
		ServiceConfig{"https://testurl"},
		&http.Client{},
		auth,
		logging.NewMockLogger(),
	)

	_, err := proxy.HandleRequest(context.Background(), Request{})

	require.Error(t, err)

	assert.ErrorContains(t, err, "auth error")
}

func TestService_HandleRequest_InvalidJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{invalid json}"))
	}))

	defer ts.Close()

	proxy := NewService(
		ServiceConfig{ts.URL},
		&http.Client{},
		&mockAuth{token: "valid"},
		logging.NewMockLogger(),
	)

	_, err := proxy.HandleRequest(context.Background(), Request{Query: "query() {}"})

	require.Error(t, err)

	assert.ErrorContains(t, err, "invalid character")
}

func TestService_HandleRequest_InvalidStatus(t *testing.T) {
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

			proxy := NewService(
				ServiceConfig{ts.URL},
				&http.Client{},
				&mockAuth{token: "valid"},
				slog.New(handler),
			)

			_, err := proxy.HandleRequest(context.Background(), Request{Query: "query() {}"})

			require.Error(t, err)

			assert.ErrorContains(
				t,
				err,
				fmt.Sprintf("expected status code 200, got %v", tt.statusCode),
			)

			assert.Len(t, handler.Entries(slog.LevelError), 1)
		})
	}
}

func TestService_HandleRequest_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	proxy := NewService(
		ServiceConfig{ts.URL},
		&http.Client{},
		&mockAuth{token: "valid-token"},
		logging.NewMockLogger(),
	)

	_, err := proxy.HandleRequest(ctx, Request{
		Query: "query() {}",
	})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
