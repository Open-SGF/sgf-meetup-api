package logging

import (
	"errors"
	"log/slog"
	"sync"
	"testing"
)

type MockErrorLogger struct {
	errors      []error
	breadcrumbs []MockBreadcrumb
	mu          sync.Mutex
}

type MockBreadcrumb struct {
	msg   string
	level slog.Level
	attrs map[string]interface{}
}

func (m *MockErrorLogger) CaptureError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, err)
}

func (m *MockErrorLogger) CaptureBreadcrumb(msg string, level slog.Level, attrs map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.breadcrumbs = append(m.breadcrumbs, MockBreadcrumb{msg, level, attrs})
}

func TestCaptureErrorHandler(t *testing.T) {
	t.Run("captures explicit error attribute", func(t *testing.T) {
		mockHandler := NewMockHandler()
		mockErrorLogger := &MockErrorLogger{}
		logger := slog.New(WithErrorLogger(mockHandler, mockErrorLogger))
		testErr := errors.New("test error")

		logger.Error("message", "error", testErr)

		if len(mockErrorLogger.errors) != 1 || !errors.Is(testErr, mockErrorLogger.errors[0]) {
			t.Errorf("Expected captured error %v, got %v", testErr, mockErrorLogger.errors)
		}

		if len(mockErrorLogger.breadcrumbs) != 1 {
			t.Fatal("Expected one breadcrumb")
		}

		bc := mockErrorLogger.breadcrumbs[0]
		if bc.msg != "message" || bc.level != slog.LevelError {
			t.Error("Incorrect breadcrumb message or level")
		}

		if mockHandler.Entries(slog.LevelError)[0].Message != "message" {
			t.Error("Original handler didn't receive message")
		}
	})

	t.Run("creates error from message when no error attribute", func(t *testing.T) {
		mockHandler := NewMockHandler()
		mockErrorLogger := &MockErrorLogger{}
		logger := slog.New(WithErrorLogger(mockHandler, mockErrorLogger))

		logger.Error("message")

		if len(mockErrorLogger.errors) != 1 || mockErrorLogger.errors[0].Error() != "message" {
			t.Error("Should create error from message")
		}
	})

	t.Run("captures errors in nested groups", func(t *testing.T) {
		mockHandler := NewMockHandler()
		mockErrorLogger := &MockErrorLogger{}
		logger := slog.New(WithErrorLogger(mockHandler, mockErrorLogger)).WithGroup("group")
		nestedErr := errors.New("nested error")

		logger.Info("message", "error", nestedErr)

		if len(mockErrorLogger.errors) != 1 || !errors.Is(nestedErr, mockErrorLogger.errors[0]) {
			t.Error("Should capture error from nested group")
		}
	})

	t.Run("captures breadcrumbs for non-error levels", func(t *testing.T) {
		mockHandler := NewMockHandler()
		mockErrorLogger := &MockErrorLogger{}
		logger := slog.New(WithErrorLogger(mockHandler, mockErrorLogger))

		logger.Info("info message")
		logger.Warn("warn message")

		if len(mockErrorLogger.breadcrumbs) != 2 {
			t.Error("Should capture breadcrumbs for all levels")
		}
	})

	t.Run("withattrs preserves error capturing", func(t *testing.T) {
		mockHandler := NewMockHandler()
		mockErrorLogger := &MockErrorLogger{}
		baseLogger := slog.New(WithErrorLogger(mockHandler, mockErrorLogger))
		logger := baseLogger.With("key", "value")
		testErr := errors.New("test error")

		logger.Error("message", "error", testErr)

		if len(mockErrorLogger.errors) != 1 {
			t.Error("WithAttrs should preserve error capturing")
		}
	})
}
