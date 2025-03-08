package apo

import "time"

type DBID string

type Uzanto struct {
	ID       DBID
	Nomo     string
	Retpoŝto string
	Pasvorto string

	Admina bool

	KreitaJa  time.Time
	ŜanĝitaJe time.Time
}

func (Uzanto) Table() string {
	return "uzantoj"
}

type Kurso struct {
	ID        DBID
	Posedanto DBID
	Nomo      string
	Kiamo     time.Time
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
	ID     DBID
	Uzanto DBID
	Kurso  DBID
}

func (Lernanto) Table() string {
	return "lernantoj"
}

type Instruisto struct {
	ID     DBID
	Uzanto DBID
	Kurso  DBID
}

func (Instruisto) Table() string {
	return "instruistoj"
}

type Hejmtasko struct {
	ID       DBID
	Lernanto DBID
	Kursero  DBID
	Teksto   string
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
	Nomo     string `json:"nomo"`
	Retpoŝŧo string `json:"retpoŝto,omitzero"`
	Pasvorto string `json:"pasvorto,omitzero"`
}

type KursoJSON struct {
	ID        DBID       `json:"id"`
	Posedanto UzantoJSON `json:"posedanto"`
	Nomo      string     `json:"nomo"`
	Kiamo     time.Time  `json:"kiamo"`
}

type KurseroJSON struct {
	ID    DBID      `json:"id"`
	Kurso KursoJSON `json:"kurso"`
	Nomo  string    `json:"nomo"`
	Kiamo time.Time `json:"kiamo"`
}

type HejmtaskoJSON struct {
	ID       DBID       `json:"id"`
	Lernanto UzantoJSON `json:"lernanto"`
	Teksto   string     `json:"teksto"`
}
