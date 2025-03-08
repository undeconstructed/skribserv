package apo

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

func hazardaID(antaŭaĵo string, longo int) DBID {
	return DBID(lib.FariHazardanID(antaŭaĵo, longo))
}

type Apo struct {
	*back
	*front
}

func Nova(db rel.Repository, log *slog.Logger) (*Apo, error) {
	lf := func(ctx context.Context) *lib.Logger {
		return lib.Log(log, ctx)
	}

	back := &back{
		db:  db,
		log: lf,
	}

	front := &front{
		back:  back,
		ident: NovaIdentilo(),
		log:   lf,
	}

	app := &Apo{
		back:  back,
		front: front,
	}

	return app, nil
}
