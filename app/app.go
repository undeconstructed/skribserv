package app

import (
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-rel/rel"
	"github.com/undeconstructed/skribserv/lib"
)

func makeRandomID(prefix string, length int) DBID {
	return DBID(lib.MakeRandomID(prefix, length))
}

type App struct {
	*back
	*front
}

func New(db rel.Repository, log *slog.Logger) (*App, error) {
	back := &back{
		db:  db,
		log: lib.SubLog(log),
	}

	front := &front{
		back:  back,
		ident: NewAuthenticator(),
		log:   lib.SubLog(log),
	}

	app := &App{
		back:  back,
		front: front,
	}

	return app, nil
}
