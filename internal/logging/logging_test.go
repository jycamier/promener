package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestVerbosityToLevel(t *testing.T) {
	tests := []struct {
		name      string
		verbosity int
		expected  slog.Level
	}{
		{"negative verbosity", -1, slog.LevelError},
		{"zero verbosity", 0, slog.LevelError},
		{"one v", 1, slog.LevelWarn},
		{"two v", 2, slog.LevelInfo},
		{"three v", 3, slog.LevelDebug},
		{"four v", 4, LevelTrace},
		{"five v", 5, LevelTrace},
		{"ten v", 10, LevelTrace},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerbosityToLevel(tt.verbosity)
			if got != tt.expected {
				t.Errorf("VerbosityToLevel(%d) = %v, want %v", tt.verbosity, got, tt.expected)
			}
		})
	}
}

func TestLoggerOutputText(t *testing.T) {
	var buf bytes.Buffer
	Init(2, &buf, "text") // INFO level, text format

	Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected log to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected log to contain 'key=value', got: %s", output)
	}
	if !strings.Contains(output, "level=INFO") {
		t.Errorf("expected log to contain 'level=INFO', got: %s", output)
	}
}

func TestLoggerOutputJSON(t *testing.T) {
	var buf bytes.Buffer
	Init(2, &buf, "json") // INFO level, JSON format

	Info("json test", "number", 42)

	output := buf.String()
	if !strings.Contains(output, `"msg":"json test"`) {
		t.Errorf("expected JSON log to contain '\"msg\":\"json test\"', got: %s", output)
	}
	if !strings.Contains(output, `"number":42`) {
		t.Errorf("expected JSON log to contain '\"number\":42', got: %s", output)
	}
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("expected JSON log to contain '\"level\":\"INFO\"', got: %s", output)
	}
}

func TestLogLevelFiltering(t *testing.T) {
	tests := []struct {
		name           string
		verbosity      int
		expectDebug    bool
		expectInfo     bool
		expectWarn     bool
		expectError    bool
	}{
		{"error only (0)", 0, false, false, false, true},
		{"warn+ (1)", 1, false, false, true, true},
		{"info+ (2)", 2, false, true, true, true},
		{"debug+ (3)", 3, true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Init(tt.verbosity, &buf, "text")

			Debug("debug message")
			Info("info message")
			Warn("warn message")
			Error("error message")

			output := buf.String()

			checkContains := func(msg string, shouldContain bool) {
				contains := strings.Contains(output, msg)
				if shouldContain && !contains {
					t.Errorf("at verbosity %d, expected output to contain '%s', got: %s", tt.verbosity, msg, output)
				}
				if !shouldContain && contains {
					t.Errorf("at verbosity %d, expected output NOT to contain '%s', got: %s", tt.verbosity, msg, output)
				}
			}

			checkContains("debug message", tt.expectDebug)
			checkContains("info message", tt.expectInfo)
			checkContains("warn message", tt.expectWarn)
			checkContains("error message", tt.expectError)
		})
	}
}

func TestTraceLevelDisplay(t *testing.T) {
	var buf bytes.Buffer
	Init(4, &buf, "text") // TRACE level

	// Log at trace level using the Trace helper
	Trace("trace message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "TRACE") {
		t.Errorf("expected TRACE level label, got: %s", output)
	}
	if !strings.Contains(output, "trace message") {
		t.Errorf("expected 'trace message', got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected 'key=value', got: %s", output)
	}
}

func TestTraceLevelJSON(t *testing.T) {
	var buf bytes.Buffer
	Init(4, &buf, "json") // TRACE level, JSON format

	Trace("trace json message")

	output := buf.String()
	if !strings.Contains(output, `"level":"TRACE"`) {
		t.Errorf("expected TRACE level in JSON, got: %s", output)
	}
	if !strings.Contains(output, `"msg":"trace json message"`) {
		t.Errorf("expected trace json message in JSON, got: %s", output)
	}
}

func TestSetVerbosity(t *testing.T) {
	var buf bytes.Buffer
	Init(0, &buf, "text") // Start at ERROR only

	// DEBUG should not appear
	Debug("should not appear")
	if strings.Contains(buf.String(), "should not appear") {
		t.Error("DEBUG should be filtered at verbosity 0")
	}

	// Change to DEBUG level
	SetVerbosity(3)
	buf.Reset()

	Debug("should appear now")
	if !strings.Contains(buf.String(), "should appear now") {
		t.Error("DEBUG should appear after SetVerbosity(3)")
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	Init(0, &buf, "text") // Start at ERROR only

	// INFO should not appear
	Info("should not appear")
	if strings.Contains(buf.String(), "should not appear") {
		t.Error("INFO should be filtered at ERROR level")
	}

	// Change to INFO level directly
	SetLevel(slog.LevelInfo)
	buf.Reset()

	Info("should appear now")
	if !strings.Contains(buf.String(), "should appear now") {
		t.Error("INFO should appear after SetLevel(slog.LevelInfo)")
	}
}

func TestWithContext(t *testing.T) {
	var buf bytes.Buffer
	Init(2, &buf, "text")

	log := With("component", "test", "version", "1.0")
	log.Info("contextual message")

	output := buf.String()
	if !strings.Contains(output, "component=test") {
		t.Errorf("expected 'component=test' in output, got: %s", output)
	}
	if !strings.Contains(output, "version=1.0") {
		t.Errorf("expected 'version=1.0' in output, got: %s", output)
	}
}

func TestDefaultLogger(t *testing.T) {
	// Reset defaultLogger to test fallback
	defaultLogger.Store(nil)

	logger := Logger()
	if logger == nil {
		t.Error("Logger() should never return nil")
	}
}
