package logging

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		require.Len(t, mockErrorLogger.errors, 1)
		assert.ErrorIs(t, testErr, mockErrorLogger.errors[0])

		require.Len(t, mockErrorLogger.breadcrumbs, 1)

		bc := mockErrorLogger.breadcrumbs[0]

		assert.Equal(t, "message", bc.msg)
		assert.Equal(t, slog.LevelError, bc.level)

		assert.Equal(t, "message", mockHandler.Entries(slog.LevelError)[0].Message)
	})

	t.Run("creates error from message when no error attribute", func(t *testing.T) {
		mockHandler := NewMockHandler()
		mockErrorLogger := &MockErrorLogger{}
		logger := slog.New(WithErrorLogger(mockHandler, mockErrorLogger))

		logger.Error("message")

		require.Len(t, mockErrorLogger.errors, 1)
		assert.Equal(t, "message", mockErrorLogger.errors[0].Error())
	})

	t.Run("captures errors in nested groups", func(t *testing.T) {
		mockHandler := NewMockHandler()
		mockErrorLogger := &MockErrorLogger{}
		logger := slog.New(WithErrorLogger(mockHandler, mockErrorLogger)).WithGroup("group")
		nestedErr := errors.New("nested error")

		logger.Info("message", "error", nestedErr)

		require.Len(t, mockErrorLogger.errors, 1)
		assert.ErrorIs(t, nestedErr, mockErrorLogger.errors[0])
	})

	t.Run("captures breadcrumbs for non-error levels", func(t *testing.T) {
		mockHandler := NewMockHandler()
		mockErrorLogger := &MockErrorLogger{}
		logger := slog.New(WithErrorLogger(mockHandler, mockErrorLogger))

		logger.Info("info message")
		logger.Warn("warn message")

		require.Len(t, mockErrorLogger.breadcrumbs, 2)
	})

	t.Run("withattrs preserves error capturing", func(t *testing.T) {
		mockHandler := NewMockHandler()
		mockErrorLogger := &MockErrorLogger{}
		baseLogger := slog.New(WithErrorLogger(mockHandler, mockErrorLogger))
		logger := baseLogger.With("key", "value")
		testErr := errors.New("test error")

		logger.Error("message", "error", testErr)

		require.Len(t, mockErrorLogger.errors, 1)
	})
}
