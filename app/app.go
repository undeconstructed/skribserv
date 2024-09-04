package app

import (
	"context"
	"net/http"

	"github.com/undeconstructed/skribserv/lib"
)

type User struct {
	ID   string
	Name string
}

type Text struct {
	ID   string
	Text string
}

type EntityResponse struct {
	Message string `json:"message"`
	Entity  any    `json:"entity"`
}

type App struct {
	users map[string]User
	texts map[string]Text
}

func New() (*App, error) {
	users := map[string]User{
		"phil": {"phil", "Phil"},
	}

	texts := map[string]Text{
		"123": {"123", "This is the text"},
	}

	return &App{
		users: users,
		texts: texts,
	}, nil
}

func (a *App) GetUser(ctx context.Context, r *http.Request) any {
	log := lib.GetLogger(r.Context())
	log.Info("get api a happening here")

	id := r.PathValue("id")
	user := a.users[id]

	return EntityResponse{
		Message: "get " + id,
		Entity:  user,
	}
}

func (a *App) GetText(ctx context.Context, r *http.Request) any {
	log := lib.GetLogger(ctx)
	log.Info("get api b happening here")

	id := r.PathValue("id")
	text := a.texts[id]

	return EntityResponse{
		Message: "get " + id,
		Entity:  text,
	}
}
