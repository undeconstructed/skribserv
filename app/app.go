package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-rel/rel"
	"github.com/go-rel/rel/where"
	"github.com/undeconstructed/skribserv/lib"
)

type EntoRespondo struct {
	Mesaĝo string `json:"mesaĝo"`
	Ento   any    `json:"ento"`
}

func log(ctx context.Context) *lib.Logger {
	return lib.DefaultLog(ctx)
}

type Apo struct {
	db rel.Repository
}

func Nova(db rel.Repository) (*Apo, error) {
	app := &Apo{
		db: db,
	}

	return app, nil
}

func (a *Apo) metiUzanton(ctx context.Context, id, pasvorto, nomo string) (*Uzanto, error) {
	if id == "" {
		id = lib.FariHazardanID("u", 5)
	}

	uzanto1 := &Uzanto{
		ID:       id,
		Nomo:     nomo,
		Pasvorto: pasvorto,
	}

	if err := a.db.Insert(ctx, uzanto1); err != nil {
		return nil, err
	}

	return uzanto1, nil
}

func (a *Apo) preniUzanton(ctx context.Context, id string) (*Uzanto, error) {
	uzanto := &Uzanto{}

	err := a.db.Find(ctx, uzanto, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return uzanto, nil
}

func (a *Apo) metiTekston(ctx context.Context, id, uzantoID, teksto string) (*Teksto, error) {
	if id == "" {
		id = lib.FariHazardanID("t", 5)
	}

	teskto1 := &Teksto{
		ID:       id,
		UzantoID: uzantoID,
		Teksto:   teksto,
	}

	if err := a.db.Insert(ctx, teskto1); err != nil {
		return nil, err
	}

	return teskto1, nil
}

func (a *Apo) preniTekston(ctx context.Context, id string) (*Teksto, error) {
	teksto := &Teksto{}

	err := a.db.Find(ctx, teksto, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return teksto, nil
}

func (a *Apo) preniTekstojnLaŭUzantoID(ctx context.Context, uzantoID string) ([]*Teksto, error) {
	var out []*Teksto

	_, err := a.db.FindAndCountAll(ctx, &out, where.Eq("uzanto_id", uzantoID))
	if err != nil {
		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return out, nil
}

func (a *Apo) Instaliĝi(mux lib.Router) {
	mux("GET /api/uzantoj", a.IdentigaMezvaro(lib.APIHandler(a.PreniUzantojn)))
	mux("GET /api/uzantoj/{id}", a.IdentigaMezvaro(lib.APIHandler(a.PreniUzanton)))
	mux("POST /api/tekstoj", a.IdentigaMezvaro(lib.APIHandler(a.SendiTekston)))
	mux("GET /api/tekstoj/", a.IdentigaMezvaro(lib.APIHandler(a.PreniTekstojn)))
	mux("GET /api/tekstoj/{id}", a.IdentigaMezvaro(lib.APIHandler(a.PreniTekston)))
}

type ctxKey int

const ctxKeyUzanto ctxKey = 1

func (a *Apo) IdentigaMezvaro(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		uID, pasvorto, ok := r.BasicAuth()
		if !ok {
			lib.SendHTTPError(w, 0, lib.ErrHTTPUnauthorized)
			return
		}

		uzanto, err := a.preniUzanton(ctx, uID)
		if err != nil {
			lib.SendHTTPError(w, 0, lib.ErrHTTPUnauthorized)
			return
		}

		if uzanto.Pasvorto != pasvorto {
			lib.SendHTTPError(w, 0, lib.ErrHTTPUnauthorized)
			return
		}

		ctx1 := context.WithValue(r.Context(), ctxKeyUzanto, uzanto)
		r1 := r.WithContext(ctx1)

		log(ctx).Debug("auth", "user", uzanto.ID)

		next(w, r1)
	}
}

func (a *Apo) preniUzantonDePeto(ctx context.Context) *Uzanto {
	return ctx.Value(ctxKeyUzanto).(*Uzanto)
}

type UzantoJSON struct {
	ID   string `json:"id"`
	Nomo string `json:"nomo"`
}

type TekstoJSON struct {
	ID       string `json:"id"`
	UzantoID string `json:"uzantoID"`
	Teksto   string `json:"teksto"`
}

func (a *Apo) PreniUzantojn(ctx context.Context, r *http.Request) any {
	return errors.New("not implemented")
}

func (a *Apo) PreniUzanton(ctx context.Context, r *http.Request) any {
	log := lib.DefaultLog(r.Context())
	log.Info("get user happening here")

	id := r.PathValue("id")

	user, err := a.preniUzanton(ctx, id)
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "get " + id,
		Ento: UzantoJSON{
			ID:   user.ID,
			Nomo: user.Nomo,
		},
	}
}

func (a *Apo) SendiTekston(ctx context.Context, r *http.Request) any {
	dec := json.NewDecoder(r.Body)

	text1 := &TekstoJSON{}
	err := dec.Decode(text1)
	if err != nil {
		return err
	}

	user := a.preniUzantonDePeto(ctx)

	if text1.UzantoID != "" && text1.UzantoID != user.ID {
		return lib.ErrHTTPForbidden
	}

	if text1.Teksto == "" {
		return fmt.Errorf("%w: missing text", lib.ErrHTTPBadRequest)
	}

	text1.UzantoID = user.ID

	text2, err := a.metiTekston(ctx, "", text1.UzantoID, text1.Teksto)
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "post",
		Ento: TekstoJSON{
			ID:       text2.ID,
			UzantoID: text2.UzantoID,
			Teksto:   text2.Teksto,
		},
	}
}

func (a *Apo) PreniTekstojn(ctx context.Context, r *http.Request) any {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		return errors.New("only querying by userID is supported")
	}

	texts, err := a.preniTekstojnLaŭUzantoID(ctx, userID)
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "get texts by " + userID,
		Ento:   texts,
	}
}

func (a *Apo) PreniTekston(ctx context.Context, r *http.Request) any {
	log := lib.DefaultLog(ctx)
	log.Info("get api b happening here")

	id := r.PathValue("id")

	text, err := a.preniTekston(ctx, id)
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "get " + id,
		Ento: TekstoJSON{
			ID:       text.ID,
			UzantoID: text.UzantoID,
			Teksto:   text.Teksto,
		},
	}
}
