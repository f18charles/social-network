// Package logger provides a structured, levelled application logger built on
// top of the standard library's log/slog package.  It is intentionally
// dependency-free so it fits the project's no-external-logging-lib policy.
//
// Usage
//
//	// In main, before starting the server:
//	logger.Init(logger.Config{Level: logger.LevelInfo, JSON: false})
//
//	// Anywhere in the application:
//	logger.Info("post created", "user_id", userID, "post_id", postID)
//	logger.Error("image upload failed", "error", err)
package logger

import (
	"context"
	"log/slog"
	"os"
)

// Level mirrors slog.Level so callers don't need to import slog directly.
type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Config controls how the global logger is initialised.
type Config struct {
	// Level is the minimum log level to emit.  Defaults to LevelInfo.
	Level Level

	// JSON emits newline-delimited JSON when true; human-readable text otherwise.
	JSON bool

	// AddSource annotates every log record with the caller file and line number.
	AddSource bool
}

// global is the package-level structured logger.
var global *slog.Logger

func init() {
	// Sensible default so callers can use the package before Init is called.
	global = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: LevelInfo,
	}))
}

// Init replaces the global logger with one built from cfg.
// Call this once from main, before starting the server.
func Init(cfg Config) {
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.JSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	global = slog.New(handler)
	slog.SetDefault(global)
}

// L returns the global structured logger for callers that need it directly.
func L() *slog.Logger { return global }

// ── Convenience wrappers ──────────────────────────────────────────────────────

// Debug logs at DEBUG level with optional key-value pairs.
func Debug(msg string, args ...any) { global.Debug(msg, args...) }

// Info logs at INFO level with optional key-value pairs.
func Info(msg string, args ...any) { global.Info(msg, args...) }

// Warn logs at WARN level with optional key-value pairs.
func Warn(msg string, args ...any) { global.Warn(msg, args...) }

// Error logs at ERROR level with optional key-value pairs.
func Error(msg string, args ...any) { global.Error(msg, args...) }

// DebugCtx logs at DEBUG level, propagating a context (useful for
// trace-ID injection when middleware populates the context).
func DebugCtx(ctx context.Context, msg string, args ...any) {
	global.DebugContext(ctx, msg, args...)
}

// InfoCtx logs at INFO level with context propagation.
func InfoCtx(ctx context.Context, msg string, args ...any) {
	global.InfoContext(ctx, msg, args...)
}

// WarnCtx logs at WARN level with context propagation.
func WarnCtx(ctx context.Context, msg string, args ...any) {
	global.WarnContext(ctx, msg, args...)
}

// ErrorCtx logs at ERROR level with context propagation.
func ErrorCtx(ctx context.Context, msg string, args ...any) {
	global.ErrorContext(ctx, msg, args...)
}

// With returns a new logger that always includes the given key-value pairs.
// Useful for per-handler loggers that carry a fixed component label.
//
//	log := logger.With("component", "post_handler")
//	log.Info("post created", "post_id", id)
func With(args ...any) *slog.Logger {
	return global.With(args...)
}
