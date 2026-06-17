package server

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// InitLogger initializes the default structured logger with the given
// level and format. Valid levels: debug, info, warn, error.
// Valid formats: text, json.
func InitLogger(level string, format string) {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     l,
		AddSource: l <= slog.LevelDebug,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.SourceKey {
				if source, ok := attr.Value.Any().(*slog.Source); ok {
					source.File = filepath.Base(source.File)
				}
			}
			return attr
		},
	}

	var h slog.Handler
	switch strings.ToLower(format) {
	case "json":
		h = slog.NewJSONHandler(os.Stderr, opts)
	default:
		h = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(h))
}

// LogRequest logs an HTTP request with method, path, status code, and
// duration. Intended for use as a Gin middleware callback.
func LogRequest(method, path string, status int, duration time.Duration) {
	slog.Info("request",
		"method", method,
		"path", path,
		"status", status,
		"duration", duration,
	)
}

// LogModelEvent logs a model lifecycle event (load, unload, evict, etc.)
// with the model name, event type, and any optional extra fields.
func LogModelEvent(model, event string, extra map[string]any) {
	args := make([]any, 0, 2+len(extra)*2)
	args = append(args, "model", model, "event", event)
	for k, v := range extra {
		args = append(args, k, v)
	}
	slog.Info("model event", args...)
}
