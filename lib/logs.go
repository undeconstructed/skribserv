package lib

import (
	"context"
	"log/slog"
	"os"
)

func MakeLogger(devMode bool) *slog.Logger {
	if devMode {
		return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

type ctxKey int

const ctxKeyLogger ctxKey = 1

func PutLogger(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, log)
}

func GetLogger(ctx context.Context) *slog.Logger {
	return ctx.Value(ctxKeyLogger).(*slog.Logger)
}
