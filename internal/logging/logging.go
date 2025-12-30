// Package logging provides structured logging using Go's standard slog package.
// It supports verbosity levels controlled by the -v flag and multiple output formats.
package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync/atomic"
)

// LevelTrace defines a custom TRACE level below DEBUG for very verbose output.
// slog levels: DEBUG=-4, INFO=0, WARN=4, ERROR=8
const LevelTrace = slog.Level(-8)

var (
	// defaultLogger is stored atomically to allow concurrent access.
	defaultLogger atomic.Pointer[slog.Logger]
	levelVar      = new(slog.LevelVar)
)

// VerbosityToLevel converts a verbosity count (0-4+) to an slog.Level.
// Mapping:
//   - 0 (no -v):  ERROR only
//   - 1 (-v):     WARN and above
//   - 2 (-vv):    INFO and above
//   - 3 (-vvv):   DEBUG and above
//   - 4+ (-vvvv): TRACE level
func VerbosityToLevel(verbosity int) slog.Level {
	switch {
	case verbosity <= 0:
		return slog.LevelError
	case verbosity == 1:
		return slog.LevelWarn
	case verbosity == 2:
		return slog.LevelInfo
	case verbosity == 3:
		return slog.LevelDebug
	default:
		return LevelTrace
	}
}

// Init initializes the global logger with the specified verbosity level and format.
// format can be "text" (human-readable) or "json" (structured).
// If output is nil, os.Stderr is used.
func Init(verbosity int, output io.Writer, format string) {
	if output == nil {
		output = os.Stderr
	}

	level := VerbosityToLevel(verbosity)
	levelVar.Set(level)

	opts := &slog.HandlerOptions{
		Level: levelVar,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize TRACE level name
			if a.Key == slog.LevelKey {
				if lvl, ok := a.Value.Any().(slog.Level); ok && lvl < slog.LevelDebug {
					a.Value = slog.StringValue("TRACE")
				}
			}
			return a
		},
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	logger := slog.New(handler)
	defaultLogger.Store(logger)
	slog.SetDefault(logger)
}

// Logger returns the configured logger instance.
// If Init() hasn't been called, returns a default logger at ERROR level.
// This function is safe for concurrent use.
func Logger() *slog.Logger {
	if logger := defaultLogger.Load(); logger != nil {
		return logger
	}
	// Return a fallback logger if Init() hasn't been called
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

// SetLevel dynamically changes the log level.
func SetLevel(level slog.Level) {
	levelVar.Set(level)
}

// SetVerbosity dynamically changes verbosity using the count (0-4+).
func SetVerbosity(verbosity int) {
	levelVar.Set(VerbosityToLevel(verbosity))
}

// Trace logs at TRACE level (below DEBUG, for very verbose output).
func Trace(msg string, args ...any) {
	Logger().Log(context.Background(), LevelTrace, msg, args...)
}

// Debug logs at DEBUG level.
func Debug(msg string, args ...any) {
	Logger().Debug(msg, args...)
}

// Info logs at INFO level.
func Info(msg string, args ...any) {
	Logger().Info(msg, args...)
}

// Warn logs at WARN level.
func Warn(msg string, args ...any) {
	Logger().Warn(msg, args...)
}

// Error logs at ERROR level.
func Error(msg string, args ...any) {
	Logger().Error(msg, args...)
}

// With returns a logger with additional context attributes.
func With(args ...any) *slog.Logger {
	return Logger().With(args...)
}
