package logging

import (
	"log/slog"
	"strings"
	"testing"
)

func TestLogTypeToString(t *testing.T) {
	tests := []struct {
		input    LogType
		expected string
	}{
		{LogTypeText, "TEXT"},
		{LogTypeJSON, "JSON"},
	}

	for _, tc := range tests {
		t.Run(tc.expected+" to string", func(t *testing.T) {
			actual := tc.input.String()

			if actual != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestParseLogType(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType LogType
		expectErr    bool
	}{
		{"text lowercase", "text", LogTypeText, false},
		{"TEXT uppercase", "TEXT", LogTypeText, false},
		{"json lowercase", "json", LogTypeJSON, false},
		{"Json mixed case", "Json", LogTypeJSON, false},
		{"json with spaces", "  json  ", LogTypeJSON, false},
		{"invalid value", "yaml", 0, true},
		{"empty string", "", 0, true},
		{"numeric input", "123", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lvl, err := ParseLogType(tc.input)

			if tc.expectErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tc.input) {
					t.Errorf("Error message %q should contain input %q", err.Error(), tc.input)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if lvl != tc.expectedType {
					t.Errorf("Expected level %v, got %v", tc.expectedType, lvl)
				}
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLvl slog.Level
		expectErr   bool
	}{
		{"debug lowercase", "debug", slog.LevelDebug, false},
		{"DEBUG uppercase", "DEBUG", slog.LevelDebug, false},
		{"info mixed case", "iNfO", slog.LevelInfo, false},
		{"warn alias", "warn", slog.LevelWarn, false},
		{"warning full", "WARNING", slog.LevelWarn, false},
		{"error with spaces", "  error  ", slog.LevelError, false},
		{"invalid level", "critical", 0, true},
		{"empty string", "", 0, true},
		{"numeric input", "123", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lvl, err := ParseLogLevel(tc.input)

			if tc.expectErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tc.input) {
					t.Errorf("Error message %q should contain input %q", err.Error(), tc.input)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if lvl != tc.expectedLvl {
					t.Errorf("Expected level %v, got %v", tc.expectedLvl, lvl)
				}
			}
		})
	}
}
