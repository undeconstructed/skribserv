package app

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-rel/rel"
	"github.com/undeconstructed/skribserv/lib"
)

// func log(ctx context.Context) *lib.Logger {
// 	return lib.DefaultLog(ctx)
// }

func makeRandomID(prefix string, length int) DBID {
	return DBID(lib.MakeRandomID(prefix, length))
}

type App struct {
	*back
	*front
}

func New(db rel.Repository, log *slog.Logger) (*App, error) {
	lf := func(ctx context.Context) *lib.Logger {
		return lib.Log(log, ctx)
	}

	back := &back{
		db:  db,
		log: lf,
	}

	front := &front{
		back:  back,
		ident: NewAuthenticator(),
		log:   lf,
	}

	app := &App{
		back:  back,
		front: front,
	}

	return app, nil
}
