package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/undeconstructed/skribserv/db"
	"github.com/undeconstructed/skribserv/lib"
)

var userType = db.Mkid("user")
var textType = db.Mkid("text")
var textByUser = db.Mkid("byUser")

type User struct {
	IDf  db.ID  `json:"-"`
	Pass string `json:"pass"`
	Name string `json:"name"`
}

func (e *User) ID() db.ID     { return e.IDf }
func (e *User) Type() db.ID   { return userType }
func (e *User) SetID(s db.ID) { e.IDf = s }

type Text struct {
	IDf    db.ID  `json:"-"`
	UserID db.ID  `json:"user_id"`
	Text   string `json:"text"`
}

func (e *Text) ID() db.ID     { return e.IDf }
func (e *Text) Type() db.ID   { return textType }
func (e *Text) SetID(s db.ID) { e.IDf = s }

type EntityResponse struct {
	Message string `json:"message"`
	Entity  any    `json:"entity"`
}

type App struct {
	db *db.DB
}

func New() (*App, error) {
	store, err := db.New("db.txt")
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	store.Index(context.Background(), &Text{}, textByUser, func(e db.Entity) db.ID {
		return e.(*Text).UserID
	})

	app := &App{
		db: store,
	}

	return app, nil
}

func log(ctx context.Context) *slog.Logger {
	return lib.GetLogger(ctx)
}

func (a *App) putUser(ctx context.Context, id, pass, name string) (*User, error) {
	if id == "" {
		id = lib.MakeRandomID("u", 5)
	}

	_, err := a.getUser(ctx, id)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			log(ctx).Error("load user", "err", err)
			return nil, lib.ErrHTTPConflict
		}
	}

	user1 := &User{
		IDf:  db.Mkid(id),
		Pass: pass,
		Name: name,
	}

	err = a.db.Store(ctx, user1)
	if err != nil {
		log(ctx).Error("store user", "err", err)
		return nil, lib.ErrHTTPConflict
	}

	return user1, nil
}

func (a *App) getUser(ctx context.Context, id string) (*User, error) {
	user0 := &User{
		IDf: db.Mkid(id),
	}

	err := a.db.Load(ctx, user0)
	if err != nil {
		return nil, err
	}

	return user0, nil
}

func (a *App) putText(ctx context.Context, id, userID, text string) (*Text, error) {
	if id == "" {
		id = lib.MakeRandomID("t", 5)
	}

	_, err := a.getText(ctx, id)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			log(ctx).Error("load user", "err", err)
			return nil, lib.ErrHTTPConflict
		}
	}

	text1 := &Text{
		IDf:    db.Mkid(id),
		UserID: db.Mkid(userID),
		Text:   text,
	}

	err = a.db.Store(ctx, text1)
	if err != nil {
		log(ctx).Error("store text", "err", err)
		return nil, lib.ErrHTTPConflict
	}

	return text1, nil
}

func (a *App) getText(ctx context.Context, id string) (*Text, error) {
	text0 := &Text{
		IDf: db.Mkid(id),
	}

	err := a.db.Load(ctx, text0)
	if err != nil {
		return nil, err
	}

	return text0, nil
}

func (a *App) getTextsByUserID(ctx context.Context, userID string) ([]Text, error) {
	var out []Text

	err := a.db.Query(ctx, textByUser, db.Mkid(userID), &out)
	if err != nil {
		return nil, err
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

		user, err := a.getUser(ctx, uID)
		if err != nil {
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
	UserID string `json:"user_id"`
	Text   string `json:"text"`
}

func (a *App) GetUsers(ctx context.Context, r *http.Request) any {
	return errors.New("not implemented")
}

func (a *App) GetUser(ctx context.Context, r *http.Request) any {
	log := lib.GetLogger(r.Context())
	log.Info("get user happening here")

	id := r.PathValue("id")

	user, err := a.getUser(ctx, id)
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "get " + id,
		Entity: UserJSON{
			ID:   user.IDf.String(),
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

	if text1.UserID != "" && text1.UserID != user.IDf.String() {
		return lib.ErrHTTPForbidden
	}

	if text1.Text == "" {
		return fmt.Errorf("%w: missing text", lib.ErrHTTPBadRequest)
	}

	text1.UserID = user.IDf.String()

	text2, err := a.putText(ctx, "", text1.UserID, text1.Text)
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "post",
		Entity: TextJSON{
			ID:     text2.IDf.String(),
			UserID: text2.UserID.String(),
			Text:   text2.Text,
		},
	}
}

func (a *App) GetTexts(ctx context.Context, r *http.Request) any {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		return fmt.Errorf("%w: only querying by userID is supported", lib.ErrHTTPBadRequest)
	}

	texts, err := a.getTextsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	out := make([]TextJSON, 0, len(texts))

	for _, text := range texts {
		out = append(out, TextJSON{
			ID:     text.IDf.String(),
			UserID: text.UserID.String(),
			Text:   text.Text,
		})
	}

	return EntityResponse{
		Message: "get texts by " + userID,
		Entity:  out,
	}
}

func (a *App) GetText(ctx context.Context, r *http.Request) any {
	log := lib.GetLogger(ctx)
	log.Info("get text happening here")

	id := r.PathValue("id")

	text, err := a.getText(ctx, id)
	if err != nil {
		return err
	}

	return EntityResponse{
		Message: "get " + id,
		Entity: TextJSON{
			ID:     text.IDf.String(),
			UserID: text.UserID.String(),
			Text:   text.Text,
		},
	}
}
