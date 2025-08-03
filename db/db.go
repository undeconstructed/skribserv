package db

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/go-rel/migration"
	"github.com/go-rel/postgres"
	"github.com/go-rel/rel"
	"github.com/undeconstructed/skribserv/db/migrations"
)

func Setup(dbdsn string, log *slog.Logger) (rel.Repository, error) {
	adapter, err := postgres.Open(dbdsn)
	if err != nil {
		return nil, err
	}

	db := rel.New(adapter)
	db.Instrumentation(func(ctx context.Context, op string, message string, args ...any) func(err error) {
		// no op for rel functions.
		if strings.HasPrefix(op, "rel-") {
			return func(error) {}
		}

		t := time.Now()

		return func(err error) {
			level := slog.LevelDebug
			if err != nil {
				level = slog.LevelError
			}

			duration := time.Since(t)

			log.Log(ctx, level, "db op", "op", op, "msg", message, "t", duration, "err", err, "args", args)
		}
	})

	err = db.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	m := migration.New(db)

	m.Register(2025010101000000, migrations.MigrateCreateInitial, migrations.RollbackCreateInitial)

	m.Migrate(context.Background())

	return db, nil
}
