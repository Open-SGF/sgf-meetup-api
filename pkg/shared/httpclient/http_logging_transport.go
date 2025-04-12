package httpclient

import (
	"log/slog"
	"net/http"
	"sgf-meetup-api/pkg/shared/clock"
	"time"
)

func NewHttpLoggingTransport(timeSource clock.TimeSource, logger *slog.Logger) http.RoundTripper {
	return &httpLoggingTransport{
		timeSource: timeSource,
		logger:     logger.WithGroup("http_client"),
	}
}

type httpLoggingTransport struct {
	timeSource clock.TimeSource
	logger     *slog.Logger
}

func (h httpLoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := h.timeSource.Now()
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		h.logger.ErrorContext(
			req.Context(),
			"request failed",
			"method", req.Method,
			"url", req.URL.String(),
			"error", err,
		)
		return resp, err
	}

	h.logger.LogAttrs(
		req.Context(),
		slog.LevelInfo,
		"request completed",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Int("status_code", resp.StatusCode),
		slog.Duration("duration", time.Since(start)),
		slog.String("content_length", resp.Header.Get("Content-Length")),
	)
	return resp, nil
}
