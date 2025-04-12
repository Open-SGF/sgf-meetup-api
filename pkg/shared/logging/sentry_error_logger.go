package logging

import (
	"github.com/getsentry/sentry-go"
	"log/slog"
)

type SentryErrorLogger struct{}

func (s *SentryErrorLogger) CaptureError(err error) {
	sentry.CaptureException(err)
}

func (s *SentryErrorLogger) CaptureBreadcrumb(msg string, level slog.Level, attrs map[string]interface{}) {
	breadcrumb := &sentry.Breadcrumb{
		Message:  msg,
		Level:    sentry.Level(level.String()),
		Data:     attrs,
		Category: "log",
	}
	sentry.AddBreadcrumb(breadcrumb)
}
