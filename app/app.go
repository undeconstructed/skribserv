package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/undeconstructed/skribserv/lib"
)

type User struct {
	ID   string
	Pass string
	Name string
}

type Text struct {
	ID     string
	UserID string
	Text   string
}

type EntityResponse struct {
	Message string `json:"message"`
	Entity  any    `json:"entity"`
}

func log(ctx context.Context) *slog.Logger {
	return lib.GetLogger(ctx)
}

type App struct {
	users map[string]*User
	texts map[string]*Text
}

func New() (*App, error) {
	users := map[string]*User{}
	texts := map[string]*Text{}

	app := &App{
		users: users,
		texts: texts,
	}

	app.putUser("phil", "pass1", "Phil")
	app.putText("", "phil", "this is the text")

	return app, nil
}

func (a *App) putUser(id, pass, name string) (*User, error) {
	if id == "" {
		id = lib.MakeRandomID("u", 5)
	}

	user0 := a.users[id]
	if user0 != nil {
		return nil, lib.ErrHTTPConflict
	}

	user1 := &User{
		ID:   id,
		Pass: pass,
		Name: name,
	}

	a.users[id] = user1

	return user1, nil
}

func (a *App) getUser(id string) (*User, error) {
	user := a.users[id]
	if user == nil {
		return nil, lib.ErrHTTPNotFound
	}

	return user, nil
}

func (a *App) putText(id, userID, text string) (*Text, error) {
	if id == "" {
		id = lib.MakeRandomID("t", 5)
	}

	text0 := a.texts[id]
	if text0 != nil {
		return nil, lib.ErrHTTPConflict
	}

	text1 := &Text{
		ID:     id,
		UserID: userID,
		Text:   text,
	}

	a.texts[id] = text1

	return text1, nil
}

func (a *App) getText(id string) (*Text, error) {
	text := a.texts[id]
	if text == nil {
		return nil, lib.ErrHTTPNotFound
	}

	return text, nil
}

func (a *App) getTextsByUserID(userID string) ([]*Text, error) {
	var out []*Text

	for _, text := range a.texts {
		if text.UserID == userID {
			out = append(out, text)
		}
	}

	return out, nil
}

func (a *App) Install(mux lib.Router) {
	mux("GET /api/users", a.AuthMiddleware(lib.APIHandler(a.GetUsers)))
	mux("GET /api/users/{id}", a.AuthMiddleware(lib.APIHandler(a.GetUser)))
	mux("POST /api/texts", a.AuthMiddleware(lib.APIHandler(a.PostText)))
	mux("GET /api/texts/", a.AuthMiddleware(lib.APIHandler(a.GetTexts)))
	mux("GET /api/texts/{id}", a.AuthMiddleware(lib.APIHandler(a.GetText)))
}

type ctxKey int

const ctxKeyUser ctxKey = 1

func (a *App) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		uID, pass, ok := r.BasicAuth()
		if !ok {
			lib.SendHTTPError(w, 0, lib.ErrHTTPUnauthorized)
			return
		}

		user, ok := a.users[uID]
		if !ok {
			lib.SendHTTPError(w, 0, lib.ErrHTTPUnauthorized)
			return
		}

		if user.Pass != pass {
			lib.SendHTTPError(w, 0, lib.ErrHTTPUnauthorized)
			return
		}

		ctx1 := context.WithValue(r.Context(), ctxKeyUser, user)
		r1 := r.WithContext(ctx1)

		log(ctx).Debug("auth", "user", user.ID)

		next(w, r1)
	}
}

func (a *App) getRequestUser(ctx context.Context) *User {
	return ctx.Value(ctxKeyUser).(*User)
}

type UserJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TextJSON struct {
	ID     string `json:"id"`
	UserID string `json:"userID"`
	Text   string `json:"text"`
}

func (a *App) GetUsers(ctx context.Context, r *http.Request) any {
	return errors.New("not implemented")
}

func (a *App) GetUser(ctx context.Context, r *http.Request) any {
	log := lib.GetLogger(r.Context())
	log.Info("get user happening here")

	id := r.PathValue("id")

	user, err := a.getUser(id)
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "get " + id,
		Entity: UserJSON{
			ID:   user.ID,
			Name: user.Name,
		},
	}
}

func (a *App) PostText(ctx context.Context, r *http.Request) any {
	dec := json.NewDecoder(r.Body)

	text1 := &TextJSON{}
	err := dec.Decode(text1)
	if err != nil {
		return err
	}

	user := a.getRequestUser(ctx)

	if text1.UserID != "" && text1.UserID != user.ID {
		return lib.ErrHTTPForbidden
	}

	if text1.Text == "" {
		return fmt.Errorf("%w: missing text", lib.ErrHTTPBadRequest)
	}

	text1.UserID = user.ID

	text2, err := a.putText("", text1.UserID, text1.Text)
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "post",
		Entity: TextJSON{
			ID:     text2.ID,
			UserID: text2.UserID,
			Text:   text2.Text,
		},
	}
}

func (a *App) GetTexts(ctx context.Context, r *http.Request) any {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		return errors.New("only querying by userID is supported")
	}

	texts, err := a.getTextsByUserID(userID)
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "get texts by " + userID,
		Entity:  texts,
	}
}

func (a *App) GetText(ctx context.Context, r *http.Request) any {
	log := lib.GetLogger(ctx)
	log.Info("get api b happening here")

	id := r.PathValue("id")

	text, err := a.getText(id)
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "get " + id,
		Entity: TextJSON{
			ID:     text.ID,
			UserID: text.UserID,
			Text:   text.Text,
		},
	}
}
