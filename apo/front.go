package apo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-rel/rel"
	"github.com/undeconstructed/skribserv/lib"
)

type front struct {
	back  *back
	ident *Identilo
	log   func(context.Context) *lib.Logger
}

func (a *front) Muntiĝi(mux lib.Router) {
	h := lib.APIHandler

	mux("POST", "/api/mi/ensaluti", h(a.Ensaluti))
	mux("POST", "/api/mi/elsaluti", h(a.Elsaluti))
	mux("GET", "/api/mi", h(a.PriMi), a.kunIdentigo)

	mux("GET", "/api/uzantoj", h(a.PreniUzantojn), a.porAdmino, a.kunIdentigo)
	mux("POST", "/api/uzantoj", h(a.SendiUzantojn), a.porAdmino, a.kunIdentigo)
	mux("GET", "/api/uzantoj/{uzanto}", h(a.PreniUzanton), a.porAdminoAŭMemo, a.kunIdentigo)

	mux("GET", "/api/kursoj", h(a.PreniKursojn), a.kunIdentigo)
	mux("POST", "/api/kursoj", h(a.SendiKursojn), a.porAdmino, a.kunIdentigo)
	mux("GET", "/api/kursoj/{kurso}", h(a.PreniKurson), a.kunIdentigo)

	mux("GET", "/api/kursoj/{kurso}/eroj", h(a.PreniKurserojn), a.kunIdentigo)
	mux("POST", "/api/kursoj/{kurso}/eroj", h(a.SendiKurserojn), a.kunIdentigo)
	mux("GET", "/api/kursoj/{kurso}/eroj/{kursero}", h(a.PreniKurseron), a.kunIdentigo)

	mux("GET", "/api/kursoj/{kurso}/eroj/{kursero}/hejmtaskoj", h(a.TroviHejmtaskojnLaŭKursero), a.kunIdentigo)

	mux("POST", "/api/kursoj/{kurso}/lernantoj", h(a.SendiLernantojn), a.porAdmino, a.kunIdentigo)
	mux("GET", "/api/kursoj/{kurso}/lernantoj/", h(a.PreniLernantojn), a.kunIdentigo)
	mux("GET", "/api/kursoj/{kurso}/lernantoj/{lernanto}", h(a.PreniLernanton), a.kunIdentigo)

	mux("GET", "/api/uzantoj/{uzanto}/kursoj", h(a.PreniKursojnDeUzanto), a.porAdminoAŭMemo, a.kunIdentigo)

	mux("POST", "/api/uzantoj/{uzanto}/hejmtaskoj", h(a.SendiHejmtaskon), a.porAdminoAŭMemo, a.kunIdentigo)
	mux("GET", "/api/uzantoj/{uzanto}/hejmtaskoj/", h(a.PreniHejmtaskojn), a.porAdminoAŭMemo, a.kunIdentigo)
	mux("GET", "/api/uzantoj/{uzanto}/hejmtaskoj/{hejmtasko}", h(a.PreniHejmtaskon), a.kunIdentigo)
}

type ctxKey int

const ctxKeyUzanto ctxKey = 1

func (a *front) kunIdentigo(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		type seancfn func() (*Uzanto, error)

		proviKuketon := func() (*Uzanto, error) {
			if seancokuketo, er := r.Cookie("Seanco"); er == nil {
				seancoID := seancokuketo.Value

				uzanto, er := a.ident.preniSeancon(seancoID)
				if er != nil {
					if errors.Is(er, ErrNeniuSeanco) {
						return nil, nil
					}

					return nil, lib.ErrHTTPUnauthorized
				}

				return uzanto, nil
			}

			return nil, nil
		}

		proviKapaĵon := func() (*Uzanto, error) {
			if retpoŝto, pasvorto, ok := r.BasicAuth(); ok {
				uzanto, er := a.back.preniUzantonPerRetpoŝto(ctx, retpoŝto)
				if er != nil {
					if errors.Is(er, rel.ErrNotFound) {
						return nil, lib.ErrHTTPUnauthorized
					}

					return nil, er
				}

				if uzanto.Pasvorto != pasvorto {
					return nil, lib.ErrHTTPUnauthorized
				}

				return uzanto, nil
			}

			return nil, nil
		}

		proviBazan := func() (*Uzanto, error) { return nil, nil }

		for _, f := range []seancfn{proviKuketon, proviKapaĵon, proviBazan} {
			uzanto, er := f()
			if er != nil {
				lib.SendHTTPError(w, 0, er)
				return
			}

			if uzanto != nil {
				ctx1 := context.WithValue(r.Context(), ctxKeyUzanto, uzanto)
				r1 := r.WithContext(ctx1)

				a.log(ctx).Debug("auth", "user", uzanto.ID)

				next(w, r1)

				return
			}
		}

		lib.SendHTTPError(w, 0, lib.ErrHTTPUnauthorized)
	}
}

func (a *front) porAdmino(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		u := a.uzantoElCtx(ctx)

		if !u.Admina {
			lib.SendHTTPError(w, 0, lib.ErrHTTPForbidden)
			return
		}

		next(w, r)
	}
}

func (a *front) porAdminoAŭMemo(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		uzantoID := r.PathValue("uzanto")

		u := a.uzantoElCtx(ctx)

		if !u.Admina && u.ID != DBID(uzantoID) {
			lib.SendHTTPError(w, 0, lib.ErrHTTPForbidden)
			return
		}

		next(w, r)
	}
}

func (a *front) Ensaluti(ctx context.Context, r *http.Request) any {
	dec := json.NewDecoder(r.Body)

	type ensalutpeto struct {
		Retpoŝto string `json:"retpoŝto"`
		Pasvorto string `json:"pasvorto"`
	}

	peto := &ensalutpeto{}
	er := dec.Decode(peto)
	if er != nil {
		return er
	}

	uzanto, er := a.back.preniUzantonPerRetpoŝto(ctx, peto.Retpoŝto)
	if er != nil {
		return er
	}

	sID, er := a.ident.metiSeancon(uzanto)
	if er != nil {
		return er
	}

	seanckuketo := &http.Cookie{
		Name:  "Seanco",
		Value: sID,
		Path:  "/",
	}

	return lib.HTTPResponse{
		Cookies: []*http.Cookie{seanckuketo},
		Data: EntoRespondo{
			Mesaĝo: "seanco " + sID,
			Ento: UzantoJSON{
				ID:   uzanto.ID,
				Nomo: uzanto.Nomo,
			},
		},
	}
}

func (a *front) Elsaluti(ctx context.Context, r *http.Request) any {
	kuketo, er := r.Cookie("Seanco")
	if er != nil {
		if errors.Is(er, http.ErrNoCookie) {
			return lib.HTTPResponse{}
		}

		return er
	}

	sID := kuketo.Value

	a.ident.forigiSeancon(sID)

	seanckuketo := &http.Cookie{
		Name:    "Seanco",
		Path:    "/",
		Expires: time.Time{},
	}

	return lib.HTTPResponse{
		Cookies: []*http.Cookie{seanckuketo},
		Data: EntoRespondo{
			Mesaĝo: "seanco " + sID,
		},
	}
}

func (a *front) uzantoElCtx(ctx context.Context) *Uzanto {
	return ctx.Value(ctxKeyUzanto).(*Uzanto)
}

func (a *front) PriMi(ctx context.Context, r *http.Request) any {
	uzanto := ctx.Value(ctxKeyUzanto).(*Uzanto)

	return EntoRespondo{
		Mesaĝo: "uzanto",
		Ento:   uzanto,
	}
}

func (a *front) PreniUzantojn(ctx context.Context, r *http.Request) any {
	uzantoj, err := a.back.preniUzantojn(ctx)
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "uzantoj",
		Ento:   uzantoj,
	}
}

func (a *front) SendiUzantojn(ctx context.Context, r *http.Request) any {
	dec := json.NewDecoder(r.Body)

	uzanto1 := &UzantoJSON{}
	err := dec.Decode(uzanto1)
	if err != nil {
		return err
	}

	uzanto2, err := a.back.metiUzanton(ctx, "", uzanto1.Nomo, uzanto1.Retpoŝŧo, uzanto1.Pasvorto)
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "nova uzanto",
		Ento: UzantoJSON{
			ID:   uzanto2.ID,
			Nomo: uzanto2.Nomo,
		},
	}
}

func (a *front) PreniUzanton(ctx context.Context, r *http.Request) any {
	uzantoID := r.PathValue("uzanto")
	if uzantoID == "" {
		return lib.ErrHTTPNotFound
	}

	user, err := a.back.preniUzanton(ctx, DBID(uzantoID))
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "uzanto " + uzantoID,
		Ento: UzantoJSON{
			ID:   user.ID,
			Nomo: user.Nomo,
		},
	}
}

func (a *front) PreniKursojn(ctx context.Context, r *http.Request) any {
	kursoj, err := a.back.preniKursojn(ctx)
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "kursoj",
		Ento:   kursoj,
	}
}

func (a *front) SendiKursojn(ctx context.Context, r *http.Request) any {
	user := a.uzantoElCtx(ctx)

	dec := json.NewDecoder(r.Body)

	kurso1 := &KursoJSON{}
	err := dec.Decode(kurso1)
	if err != nil {
		return err
	}

	posedanto := kurso1.Posedanto.ID
	if posedanto == "" {
		posedanto = user.ID
	}

	kurso2, err := a.back.metiKurson(ctx, "", posedanto, kurso1.Nomo, kurso1.Kiamo)
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "nova kurso",
		Ento: KursoJSON{
			ID:   kurso2.ID,
			Nomo: kurso2.Nomo,
			Posedanto: UzantoJSON{
				ID: user.ID,
			},
		},
	}
}

func (a *front) PreniKurson(ctx context.Context, r *http.Request) any {
	kursoID := r.PathValue("kurso")
	if kursoID == "" {
		return lib.ErrHTTPNotFound
	}

	kurso, err := a.back.preniKurson(ctx, DBID(kursoID))
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "kurso",
		Ento: KursoJSON{
			ID:   kurso.ID,
			Nomo: kurso.Nomo,
			Posedanto: UzantoJSON{
				ID:   kurso.PosedantoX.ID,
				Nomo: kurso.PosedantoX.Nomo,
			},
		},
	}
}

func (a *front) PreniKurserojn(ctx context.Context, r *http.Request) any {
	kursoID := r.PathValue("kurso")
	if kursoID == "" {
		return lib.ErrHTTPNotFound
	}

	kurseroj, err := a.back.preniKurserojnLaŭKurso(ctx, DBID(kursoID))
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "kurseroj de " + kursoID,
		Ento:   kurseroj,
	}
}

func (a *front) SendiKurserojn(ctx context.Context, r *http.Request) any {
	kursoID := r.PathValue("kurso")
	if kursoID == "" {
		return lib.ErrHTTPNotFound
	}

	uzanto := a.uzantoElCtx(ctx)

	dec := json.NewDecoder(r.Body)

	kursero1 := &KurseroJSON{}
	er := dec.Decode(kursero1)
	if er != nil {
		return er
	}

	if kursero1.Kurso.ID != "" && kursero1.Kurso.ID != DBID(kursoID) {
		return lib.ErrHTTPBadRequest
	}

	kurso, er := a.back.preniKurson(ctx, DBID(kursoID))
	if er != nil {
		return er
	}

	if kurso.PosedantoID != uzanto.ID {
		return lib.ErrHTTPForbidden
	}

	kursero2, er := a.back.metiKurseron(ctx, "", kurso.ID, kursero1.Nomo, kursero1.Kiamo)
	if er != nil {
		return er
	}

	return EntoRespondo{
		Mesaĝo: "nova kursero",
		Ento: KurseroJSON{
			ID:   kursero2.ID,
			Nomo: kursero2.Nomo,
			Kurso: KursoJSON{
				ID: kurso.ID,
			},
		},
	}
}

func (a *front) PreniKurseron(ctx context.Context, r *http.Request) any {
	return ErrNerealigite
}

func (a *front) PreniHejmtaskojn(ctx context.Context, r *http.Request) any {
	uzantoID := r.PathValue("uzanto")
	if uzantoID == "" {
		return lib.ErrHTTPNotFound
	}

	hejmtaskoj, err := a.back.preniHejmtaskojnLaŭUzanto(ctx, DBID(uzantoID))
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "hejmtaskoj de " + uzantoID,
		Ento:   hejmtaskoj,
	}
}

func (a *front) PreniLernantojn(ctx context.Context, r *http.Request) any {
	kursoID := r.PathValue("kurso")
	if kursoID == "" {
		return lib.ErrHTTPNotFound
	}

	lernantoj, err := a.back.preniLernantojnLaŭKurso(ctx, DBID(kursoID))
	if err != nil {
		return err
	}

	out := make([]LernantoJSON, 0, len(lernantoj))

	for _, l := range lernantoj {
		out = append(out, LernantoJSON{
			ID: l.ID,
			Uzanto: UzantoJSON{
				ID: l.UzantoID,
			},
		})
	}

	return EntoRespondo{
		Mesaĝo: "lernantoj de " + kursoID,
		Ento:   out,
	}
}

func (a *front) SendiLernantojn(ctx context.Context, r *http.Request) any {
	kursoID := r.PathValue("kurso")
	if kursoID == "" {
		return lib.ErrHTTPNotFound
	}

	dec := json.NewDecoder(r.Body)

	lernanto1 := &LernantoJSON{}
	er := dec.Decode(lernanto1)
	if er != nil {
		return er
	}

	if lernanto1.Kurso.ID != "" && lernanto1.Kurso.ID != DBID(kursoID) {
		return lib.ErrHTTPBadRequest
	}

	lernanto2, er := a.back.metiLernanton(ctx, "", lernanto1.Uzanto.ID, DBID(kursoID))
	if er != nil {
		return er
	}

	return EntoRespondo{
		Mesaĝo: "nova lernanto",
		Ento: LernantoJSON{
			ID: lernanto2.ID,
			Uzanto: UzantoJSON{
				ID: lernanto2.UzantoID,
			},
			Kurso: KursoJSON{
				ID: lernanto2.KursoID,
			},
		},
	}
}

func (a *front) PreniLernanton(ctx context.Context, r *http.Request) any {
	return ErrNerealigite
}

func (a *front) PreniKursojnDeUzanto(ctx context.Context, r *http.Request) any {
	uzantoID := r.PathValue("uzanto")
	if uzantoID == "" {
		return lib.ErrHTTPNotFound
	}

	kursoj, err := a.back.preniKursojnLaŭUzanto(ctx, DBID(uzantoID))
	if err != nil {
		return err
	}

	out := make([]KursoJSON, 0, len(kursoj))

	for _, l := range kursoj {
		out = append(out, KursoJSON{
			ID:    l.KursoID,
			Nomo:  l.KursoX.Nomo,
			Kiamo: l.KursoX.Kiamo,
		})
	}

	return EntoRespondo{
		Mesaĝo: "kursoj de " + uzantoID,
		Ento:   out,
	}
}

func (a *front) SendiHejmtaskon(ctx context.Context, r *http.Request) any {
	uzantoID := r.PathValue("uzanto")
	if uzantoID == "" {
		return lib.ErrHTTPNotFound
	}

	uzanto := a.uzantoElCtx(ctx)

	if uzantoID != string(uzanto.ID) {
		return lib.ErrHTTPForbidden
	}

	dec := json.NewDecoder(r.Body)

	hejmtasko1 := &HejmtaskoJSON{}
	err := dec.Decode(hejmtasko1)
	if err != nil {
		return err
	}

	if hejmtasko1.Lernanto.ID != "" && hejmtasko1.Lernanto.ID != uzanto.ID {
		return lib.ErrHTTPForbidden
	}

	if hejmtasko1.Teksto == "" {
		return fmt.Errorf("%w: mankas teksto", lib.ErrHTTPBadRequest)
	}

	lernanto := uzanto.ID

	hejmtasko, err := a.back.metiHejmtaskon(ctx, "", lernanto, hejmtasko1.Teksto)
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "nova hejmtasko",
		Ento: HejmtaskoJSON{
			ID: hejmtasko.ID,
			Lernanto: UzantoJSON{
				ID: hejmtasko.LernantoID,
			},
			Teksto: hejmtasko.Teksto,
		},
	}
}

func (a *front) PreniHejmtaskon(ctx context.Context, r *http.Request) any {
	hejmtaskoID := r.PathValue("hejmtasko")

	hejmtasko, err := a.back.preniHejmtaskon(ctx, DBID(hejmtaskoID))
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "hejmtasko " + hejmtaskoID,
		Ento: HejmtaskoJSON{
			ID: hejmtasko.ID,
			Lernanto: UzantoJSON{
				ID: hejmtasko.LernantoID,
			},
		},
	}
}

func (a *front) TroviHejmtaskojnLaŭKursero(ctx context.Context, r *http.Request) any {
	kurso, kursero := r.PathValue("kurso"), r.PathValue("kursero")
	if kurso == "" || kursero == "" {
		return lib.ErrHTTPNotFound
	}

	hejmtaskoj, err := a.back.preniHejmtaskojnLaŭKursero(ctx, DBID(kurso), DBID(kursero))
	if err != nil {
		return err
	}

	return EntoRespondo{
		Mesaĝo: "hejmtaskoj pri " + kursero,
		Ento:   hejmtaskoj,
	}
}
