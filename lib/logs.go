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

// WithLogValues would merge the log fields from contexts if that were possible
// func WithLogValues(ctx context.Context, ctx1 context.Context) context.Context {
// }

// Logger provides an easy way to access a logger with a particular context built in.
// This allows to call the short names e.g. `Debug` instead of `DebugContext`.
type Logger struct {
	raw *slog.Logger
	ctx context.Context
}

func Log(log *slog.Logger, ctx context.Context) *Logger {
	return &Logger{log, ctx}
}

func DefaultLog(ctx context.Context) *Logger {
	return Log(slog.Default(), ctx)
}

func (log *Logger) Raw() *slog.Logger {
	return log.raw
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
