package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     Level
		logLevel  Level
		shouldLog bool
	}{
		{"debug below info", LevelDebug, LevelInfo, false},
		{"info at info", LevelInfo, LevelInfo, true},
		{"warn at info", LevelWarn, LevelInfo, true},
		{"error at info", LevelError, LevelInfo, true},
		{"debug at debug", LevelDebug, LevelDebug, true},
		{"error at debug", LevelError, LevelDebug, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := &Logger{
				level:    tt.logLevel,
				writer:   &buf,
				useColor: false,
			}

			l.log(tt.level, "test message")

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldLog {
				t.Errorf("expected shouldLog=%v, got output=%v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestLoggerFormat(t *testing.T) {
	var buf bytes.Buffer
	l := &Logger{
		module:   "test",
		level:    LevelDebug,
		writer:   &buf,
		useColor: false,
	}

	l.Info("hello world", F("key", "value"))

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("output should contain INFO level")
	}
	if !strings.Contains(output, "hello world") {
		t.Error("output should contain message")
	}
	if !strings.Contains(output, "[test]") {
		t.Error("output should contain module name")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("output should contain field")
	}
}

func TestLoggerWithModule(t *testing.T) {
	var buf bytes.Buffer
	l := &Logger{
		module:   "parent",
		level:    LevelInfo,
		writer:   &buf,
		useColor: false,
	}

	child := l.WithModule("child")
	child.Info("child message")

	output := buf.String()
	if !strings.Contains(output, "[child]") {
		t.Error("child logger should have child module name")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input  string
		expect Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"unknown", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLevel(tt.input)
			if got != tt.expect {
				t.Errorf("ParseLevel(%s) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestLevelString(t *testing.T) {
	if LevelInfo.String() != "INFO" {
		t.Errorf("LevelInfo.String() = %s, want INFO", LevelInfo.String())
	}
}
