package apo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-rel/rel"
	"github.com/go-rel/rel/where"
	"github.com/undeconstructed/skribserv/lib"
)

type back struct {
	db  rel.Repository
	log func(context.Context) *lib.Logger
}

func (a *back) preniUzantojn(ctx context.Context) ([]*Uzanto, error) {
	var out []*Uzanto

	err := a.db.FindAll(ctx, &out)
	if err != nil {
		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return out, nil
}

func (a *back) metiUzanton(ctx context.Context, id DBID, nomo, retpoŝŧo, pasvorto string) (*Uzanto, error) {
	if id == "" {
		id = hazardaID("u", 5)
	}

	uzanto1 := &Uzanto{
		ID:       id,
		Nomo:     nomo,
		Retpoŝto: retpoŝŧo,
		Pasvorto: pasvorto,
	}

	if err := a.db.Insert(ctx, uzanto1); err != nil {
		return nil, err
	}

	return uzanto1, nil
}

func (a *back) preniUzanton(ctx context.Context, id DBID) (*Uzanto, error) {
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

func (a *back) preniUzantonPerRetpoŝto(ctx context.Context, retpoŝto string) (*Uzanto, error) {
	uzanto := &Uzanto{}

	err := a.db.Find(ctx, uzanto, where.Eq("retpoŝto", retpoŝto))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return uzanto, nil
}

func (a *back) preniKursojn(ctx context.Context) ([]*Kurso, error) {
	var out []*Kurso

	err := a.db.FindAll(ctx, &out)
	if err != nil {
		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return out, nil
}

func (a *back) metiKurson(ctx context.Context, id DBID, posedanto DBID, nomo string, kiamo time.Time) (*Kurso, error) {
	if id == "" {
		id = hazardaID("k", 5)
	}

	kurso1 := &Kurso{
		ID:        id,
		Nomo:      nomo,
		Posedanto: posedanto,
		Kiamo:     kiamo,
	}

	if err := a.db.Insert(ctx, kurso1); err != nil {
		return nil, err
	}

	return kurso1, nil
}

func (a *back) preniKurson(ctx context.Context, id DBID) (*Kurso, error) {
	kurso := &Kurso{}

	err := a.db.Find(ctx, kurso, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return kurso, nil
}

func (a *back) metiLernanton(ctx context.Context, id, uzanto, kurso DBID) (*Lernanto, error) {
	if id == "" {
		id = hazardaID("u", 5)
	}

	lernanto1 := &Lernanto{
		ID:     id,
		Uzanto: uzanto,
		Kurso:  kurso,
	}

	if err := a.db.Insert(ctx, lernanto1); err != nil {
		return nil, err
	}

	return lernanto1, nil
}

func (a *back) preniLernantojnLaŭKurso(ctx context.Context, kurso DBID) ([]*Lernanto, error) {
	var out []*Lernanto

	err := a.db.FindAll(ctx, out, where.Eq("kurso", kurso))
	if err != nil {
		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return out, nil
}

func (a *back) preniLernantojnLaŭUzanto(ctx context.Context, uzanto DBID) ([]*Lernanto, error) {
	var out []*Lernanto

	err := a.db.FindAll(ctx, out, where.Eq("uzanto", uzanto))
	if err != nil {
		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return out, nil
}

func (a *back) preniKurserojn(ctx context.Context, kurso DBID) ([]*Kursero, error) {
	var out []*Kursero

	err := a.db.FindAll(ctx, &out, where.Eq("kurso", kurso))
	if err != nil {
		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return out, nil
}

func (a *back) metiKurseron(ctx context.Context, id, kurso DBID, nomo string, kiamo time.Time) (*Kursero, error) {
	if id == "" {
		id = hazardaID("ke", 5)
	}

	kursero1 := &Kursero{
		ID:    id,
		Kurso: kurso,
		Nomo:  nomo,
		Kiamo: kiamo,
	}

	if err := a.db.Insert(ctx, kursero1); err != nil {
		return nil, err
	}

	return kursero1, nil
}

func (a *back) preniKurseron(ctx context.Context, id DBID) (*Kursero, error) {
	kursero := &Kursero{}

	err := a.db.Find(ctx, kursero, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return kursero, nil
}

func (a *back) preniKurserojnLaŭKurso(ctx context.Context, kurso DBID) ([]*Kursero, error) {
	var out []*Kursero

	err := a.db.FindAll(ctx, out, where.Eq("kurso", kurso), rel.SortDesc("kiam"))
	if err != nil {
		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return out, nil
}

func (a *back) metiHejmtaskon(ctx context.Context, id, lernanto DBID, teksto string) (*Hejmtasko, error) {
	if id == "" {
		id = hazardaID("t", 5)
	}

	teskto1 := &Hejmtasko{
		ID:       id,
		Lernanto: lernanto,
		Teksto:   teksto,
	}

	if err := a.db.Insert(ctx, teskto1); err != nil {
		return nil, err
	}

	return teskto1, nil
}

func (a *back) preniHejmtaskojnLaŭUzanto(ctx context.Context, uzantoID DBID) ([]*Hejmtasko, error) {
	var out []*Hejmtasko

	err := a.db.FindAll(ctx, &out, where.Eq("uzanto", uzantoID))
	if err != nil {
		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return out, nil
}

func (a *back) preniHejmtaskojnLaŭKursero(ctx context.Context, kurso, kursero DBID) ([]*Hejmtasko, error) {
	var out []*Hejmtasko

	err := a.db.FindAll(ctx, &out, where.Eq("kursero", kursero))
	if err != nil {
		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return out, nil
}

func (a *back) preniHejmtaskon(ctx context.Context, id DBID) (*Hejmtasko, error) {
	teksto := &Hejmtasko{}

	err := a.db.Find(ctx, teksto, where.Eq("id", id))
	if err != nil {
		if errors.Is(err, rel.ErrNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("db (legi): %w", err)
	}

	return teksto, nil
}
