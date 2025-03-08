package apo

import (
	"sync"

	"github.com/undeconstructed/skribserv/lib"
)

type Seanco struct {
	uzanto *Uzanto
}

type Identilo struct {
	sync.RWMutex
	seancoj map[string]*Uzanto
}

func NovaIdentilo() *Identilo {
	return &Identilo{
		seancoj: map[string]*Uzanto{},
	}
}

func (ai *Identilo) metiSeancon(uzanto *Uzanto) (string, error) {
	id := lib.FariHazardanID("seanco", 10)

	ai.Lock()
	defer ai.Unlock()

	ai.seancoj[id] = uzanto

	return id, nil
}

func (ai *Identilo) preniSeancon(id string) (*Uzanto, error) {
	ai.Lock()
	defer ai.Unlock()

	u, okej := ai.seancoj[id]
	if !okej {
		return nil, ErrNeniuSeanco
	}

	return u, nil
}
