package logging

import (
	"log/slog"
	"sync"
	"testing"
	"time"
)

func TestMockLogger(t *testing.T) {
	t.Run("basic logging", func(t *testing.T) {
		handler := NewMockHandler()
		logger := slog.New(handler)

		logger.Info("test message", "user", "john", "age", 30)
		logger.Warn("warning message", "error", "something wrong")

		entries := handler.AllEntries()
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}

		infoEntry := entries[0]
		if infoEntry.Level != slog.LevelInfo || infoEntry.Message != "test message" {
			t.Error("incorrect info log entry")
		}
		if infoEntry.Attrs["user"].(string) != "john" || infoEntry.Attrs["age"].(int64) != 30 {
			t.Error("missing info log attributes")
		}

		warnEntry := entries[1]
		if warnEntry.Level != slog.LevelWarn || warnEntry.Message != "warning message" {
			t.Error("incorrect warn log entry")
		}
	})

	t.Run("with attributes", func(t *testing.T) {
		handler := NewMockHandler()
		logger := slog.New(handler).With("service", "auth", "version", 1.2)

		logger.Error("failed request", "path", "/login", "status", 500)

		entries := handler.AllEntries()
		if len(entries) != 1 {
			t.Fatal("expected 1 entry")
		}

		verifyAttributes(t, entries[0].Attrs, map[string]any{
			"service": "auth",
			"version": 1.2,
			"path":    "/login",
			"status":  int64(500),
		})
	})

	t.Run("with group and attributes", func(t *testing.T) {
		handler := NewMockHandler()
		logger := slog.New(handler).WithGroup("request").With("method", "GET")

		logger.Debug("processing request", "path", "/api", "duration", 150*time.Millisecond)

		entries := handler.AllEntries()
		verifyAttributes(t, entries[0].Attrs, map[string]any{
			"request.method":   "GET",
			"request.path":     "/api",
			"request.duration": 150 * time.Millisecond,
		})
	})

	t.Run("with nested group and attributes", func(t *testing.T) {
		handler := NewMockHandler()
		logger := slog.New(handler).WithGroup("service").With("name", "serviceName")

		logger.Debug("starting service", "someAttr", "someValue")
		logger.Debug("starting service", "otherAttr", "otherValue")

		nestedLogger := logger.WithGroup("repository").With("name", "repositoryName")

		nestedLogger.Debug("starting repository", "someAttr", "someValue")

		entries := handler.AllEntries()
		verifyAttributes(t, entries[0].Attrs, map[string]any{
			"service.name":     "serviceName",
			"service.someAttr": "someValue",
		})

		verifyAttributes(t, entries[1].Attrs, map[string]any{
			"service.name":      "serviceName",
			"service.otherAttr": "otherValue",
		})

		verifyAttributes(t, entries[2].Attrs, map[string]any{
			"service.name":                "serviceName",
			"service.repository.name":     "repositoryName",
			"service.repository.someAttr": "someValue",
		})
	})

	t.Run("concurrent logging", func(t *testing.T) {
		handler := NewMockHandler()
		logger := slog.New(handler)
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				logger.Info("concurrent message", "worker", n)
			}(i)
		}

		wg.Wait()
		entries := handler.AllEntries()
		if len(entries) != 100 {
			t.Errorf("expected 100 entries, got %d", len(entries))
		}
	})

	t.Run("reset", func(t *testing.T) {
		handler := NewMockHandler()
		logger := slog.New(handler)

		logger.Info("first message")
		handler.Reset()
		logger.Info("second message")

		entries := handler.AllEntries()
		if len(entries) != 1 || entries[0].Message != "second message" {
			t.Error("reset didn't clear entries properly")
		}
	})

	t.Run("filter entries by level", func(t *testing.T) {
		handler := NewMockHandler()
		logger := slog.New(handler)

		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warning message")

		infoEntries := handler.Entries(slog.LevelInfo)
		if len(infoEntries) != 1 || infoEntries[0].Message != "info message" {
			t.Error("failed to filter info entries")
		}

		warnEntries := handler.Entries(slog.LevelWarn)
		if len(warnEntries) != 1 || warnEntries[0].Message != "warning message" {
			t.Error("failed to filter warn entries")
		}
	})
}

func verifyAttributes(t *testing.T, actual map[string]any, expected map[string]any) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("attribute count mismatch: got %d, want %d", len(actual), len(expected))
	}

	for k, v := range expected {
		actualVal, exists := actual[k]
		if !exists {
			t.Errorf("missing attribute %q", k)
			continue
		}
		if actualVal != v {
			t.Errorf("attribute %q: got %v, want %v", k, actualVal, v)
		}
	}
}
