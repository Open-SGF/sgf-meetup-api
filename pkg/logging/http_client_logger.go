package logging

import (
	"log/slog"
	"net/http"
	"time"
)

func NewHttpLoggingTransport(logger *slog.Logger) http.RoundTripper {
	return &httpLoggingTransport{
		logger: logger.WithGroup("http_client"),
	}
}

type httpLoggingTransport struct {
	logger *slog.Logger
}

func (h httpLoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
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
