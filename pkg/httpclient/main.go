package httpclient

import (
	"log/slog"
	"net/http"
	"sgf-meetup-api/pkg/clock"
)

func DefaultClient(timeSource clock.TimeSource, logger *slog.Logger) *http.Client {
	return &http.Client{Transport: NewHttpLoggingTransport(timeSource, logger)}
}
