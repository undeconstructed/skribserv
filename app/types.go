package app

import "time"

type Uzanto struct {
	ID       string
	Nomo     string
	Respoŝto string
	Pasvorto string

	KreitaJa  time.Time
	ŜanĝitaJe time.Time
}

type Teksto struct {
	ID       string
	UzantoID string
	Teksto   string

	KreitaJa  time.Time
	ŜanĝitaJe time.Time
}
