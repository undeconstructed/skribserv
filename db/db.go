package db

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-rel/migration"
	"github.com/go-rel/postgres"
	"github.com/go-rel/rel"
	"github.com/undeconstructed/skribserv/db/migrations"
)

func Munti(dbdsn string) (rel.Repository, error) {
	adapter, err := postgres.Open(dbdsn)
	if err != nil {
		return nil, err
	}

	db := rel.New(adapter)

	err = db.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	m := migration.New(db)

	m.Register(2025010101000000, migrations.MigrateCreateInitial, migrations.RollbackCreateInitial)

	m.Migrate(context.Background())

	return db, nil
}
