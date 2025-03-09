package apo

import "time"

type DBID string

type Uzanto struct {
	ID       DBID
	Nomo     string
	Retpoŝto string
	Pasvorto string

	Admina bool

	KreitaJe  time.Time
	ŜanĝitaJe time.Time
}

func (Uzanto) Table() string {
	return "uzantoj"
}

type Kurso struct {
	ID          DBID
	PosedantoID DBID   `db:"posedanto"`
	PosedantoX  Uzanto `ref:"posedanto" fk:"id"`
	Nomo        string
	Kiamo       time.Time

	Kurseroj []Kursero `ref:"id" fk:"kurso"`
}

func (Kurso) Table() string {
	return "kursoj"
}

type Kursero struct {
	ID    DBID
	Kurso DBID
	Nomo  string
	Kiamo time.Time
}

func (Kursero) Table() string {
	return "kurseroj"
}

type Lernanto struct {
	ID       DBID
	UzantoID DBID   `db:"uzanto"`
	UzantoX  Uzanto `ref:"uzanto" fk:"id"`
	KursoID  DBID   `db:"kurso"`
	KursoX   Kurso  `ref:"kurso" fk:"id"`
}

func (Lernanto) Table() string {
	return "lernantoj"
}

type Instruisto struct {
	ID       DBID
	UzantoID DBID   `db:"uzanto"`
	UzantoX  Uzanto `ref:"uzanto" fk:"id"`
	KursoID  DBID   `db:"kurso"`
}

func (Instruisto) Table() string {
	return "instruistoj"
}

type Hejmtasko struct {
	ID         DBID
	LernantoID DBID    `db:"lernanto"`
	KurseroID  DBID    `db:"kursero"`
	KurseroX   Kursero `ref:"kursero" fk:"id"`
	Teksto     string
}

func (Hejmtasko) Table() string {
	return "hejmtaskoj"
}

type EntoRespondo struct {
	Mesaĝo string `json:"mesaĝo"`
	Ento   any    `json:"ento"`
}

type UzantoJSON struct {
	ID       DBID   `json:"id"`
	Nomo     string `json:"nomo,omitzero"`
	Retpoŝŧo string `json:"retpoŝto,omitzero"`
	Pasvorto string `json:"pasvorto,omitzero"`
}

type KursoJSON struct {
	ID        DBID       `json:"id"`
	Posedanto UzantoJSON `json:"posedanto,omitzero"`
	Nomo      string     `json:"nomo,omitzero"`
	Kiamo     time.Time  `json:"kiamo,omitzero"`
}

type KurseroJSON struct {
	ID    DBID      `json:"id"`
	Kurso KursoJSON `json:"kurso,omitzero"`
	Nomo  string    `json:"nomo,omitzero"`
	Kiamo time.Time `json:"kiamo,omitzero"`
}

type LernantoJSON struct {
	ID     DBID       `json:"id"`
	Kurso  KursoJSON  `json:"kurso,omitzero"`
	Uzanto UzantoJSON `json:"uzanto,omitzero"`
}

type HejmtaskoJSON struct {
	ID       DBID       `json:"id"`
	Lernanto UzantoJSON `json:"lernanto,omitzero"`
	Teksto   string     `json:"teksto,omitzero"`
}
