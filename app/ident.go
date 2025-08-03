package app

import (
	"sync"

	"github.com/undeconstructed/skribserv/lib"
)

type Session struct {
	user *User
}

type Authenticator struct {
	sync.RWMutex
	sessions map[string]*Session
}

func NewAuthenticator() *Authenticator {
	return &Authenticator{
		sessions: map[string]*Session{},
	}
}

func (ai *Authenticator) putSession(user *User) (string, error) {
	id := lib.MakeRandomID("seanco", 10)

	ai.Lock()
	defer ai.Unlock()

	ai.sessions[id] = &Session{user: user}

	return id, nil
}

func (ai *Authenticator) getSessionUser(id string) (*User, error) {
	ai.Lock()
	defer ai.Unlock()

	u, ok := ai.sessions[id]
	if !ok {
		return nil, ErrNoSession
	}

	return u.user, nil
}

func (ai *Authenticator) deleteSession(id string) {
	ai.Lock()
	defer ai.Unlock()

	delete(ai.sessions, id)
}
