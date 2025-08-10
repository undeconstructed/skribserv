package lib

import (
	"context"
	"log/slog"
	"os"

	slogcontext "github.com/PumpkinSeed/slog-context"
	"github.com/phsym/console-slog"
)

// Make creates a new logger with some default values.
func MakeLogger(level slog.Level, devMode bool) *slog.Logger {
	if devMode {
		return slog.New(slogcontext.NewHandler(console.NewHandler(os.Stderr, &console.HandlerOptions{Level: level})))
	}

	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// WithValue helps with putting log values into a context.
func WithLogValue(parent context.Context, key string, val any) context.Context {
	return slogcontext.WithValue(parent, key, val)
}

// MakeContextLogger can apply a context to an existing logger
type MakeContextLogger func(context.Context) *Logger

// Logger provides an easy way to access a logger with a particular context built in.
// This allows to call the short names e.g. `Debug` instead of `DebugContext`.
type Logger struct {
	raw *slog.Logger
	ctx context.Context
}

// Log wraps an [slog.Logger] and a context, so that all logging will use that context.
func Log(log *slog.Logger, ctx context.Context) *Logger {
	return &Logger{log, ctx}
}

// SubLog makes a [MakeLogger] function.
func SubLog(log *slog.Logger) MakeContextLogger {
	return func(ctx context.Context) *Logger {
		return Log(log, ctx)
	}
}

// DefaultLog is [Log] with [slog.Default].
func DefaultLog(ctx context.Context) *Logger {
	return Log(slog.Default(), ctx)
}

// Raw strips the context, going back to an [slog.Logger].
func (log *Logger) Raw() *slog.Logger {
	return log.raw
}

// Sub is a [MakeContextLogger] function, based on the underlying [slog.Logger].
func (log *Logger) Sub(ctx context.Context) *Logger {
	return Log(log.raw, ctx)
}

func (log *Logger) Debug(msg string, args ...any) {
	if log.ctx != nil {
		log.raw.DebugContext(log.ctx, msg, args...)
	} else {
		log.raw.Debug(msg, args...)
	}
}

func (log *Logger) Info(msg string, args ...any) {
	if log.ctx != nil {
		log.raw.InfoContext(log.ctx, msg, args...)
	} else {
		log.raw.Info(msg, args...)
	}
}

func (log *Logger) Warn(msg string, args ...any) {
	if log.ctx != nil {
		log.raw.WarnContext(log.ctx, msg, args...)
	} else {
		log.raw.Warn(msg, args...)
	}
}

func (log *Logger) Error(msg string, args ...any) {
	if log.ctx != nil {
		log.raw.ErrorContext(log.ctx, msg, args...)
	} else {
		log.raw.Error(msg, args...)
	}
}
