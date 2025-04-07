package logging

import (
	"context"
	"errors"
	"log/slog"
)

type ErrorLogger interface {
	CaptureError(err error)
	CaptureBreadcrumb(msg string, level slog.Level, attrs map[string]interface{})
}

func WithErrorLogger(handler slog.Handler, errorLogger ErrorLogger) slog.Handler {
	return &captureErrorHandler{
		handler:     handler,
		errorLogger: errorLogger,
	}
}

type captureErrorHandler struct {
	handler     slog.Handler
	errorLogger ErrorLogger
}

func (h *captureErrorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *captureErrorHandler) Handle(ctx context.Context, record slog.Record) error {
	attrs := h.collectAttrs(record)
	h.errorLogger.CaptureBreadcrumb(record.Message, record.Level, attrs)

	var primaryErr error
	record.Attrs(func(attr slog.Attr) bool {
		if primaryErr == nil {
			h.getErrorFromAttr(attr, &primaryErr)
		}
		return true
	})

	if primaryErr == nil && record.Level >= slog.LevelError {
		primaryErr = errors.New(record.Message)
	}

	if primaryErr != nil {
		h.errorLogger.CaptureError(primaryErr)
	}

	return h.handler.Handle(ctx, record)
}

func (h *captureErrorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &captureErrorHandler{
		handler:     h.handler.WithAttrs(attrs),
		errorLogger: h.errorLogger,
	}
}

func (h *captureErrorHandler) WithGroup(name string) slog.Handler {
	return &captureErrorHandler{
		handler:     h.handler.WithGroup(name),
		errorLogger: h.errorLogger,
	}
}

func (h *captureErrorHandler) getErrorFromAttr(attr slog.Attr, err *error) {
	if attr.Value.Kind() == slog.KindGroup {
		for _, groupAttr := range attr.Value.Group() {
			h.getErrorFromAttr(groupAttr, err)
		}
	} else if *err == nil {
		if e, ok := attr.Value.Any().(error); ok {
			*err = e
		}
	}
}

func (h *captureErrorHandler) collectAttrs(record slog.Record) map[string]interface{} {
	attrs := make(map[string]interface{})
	record.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()
		return true
	})
	return attrs
}
