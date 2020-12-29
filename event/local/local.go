package local

import (
	"errors"
	"log"
	"sync"

	"github.com/akarasz/pajthy-backend/event"
)

type Internal struct {
	sync.Mutex
	sessions map[string]*session
}

func New() *Internal {
	return &Internal{
		sessions: map[string]*session{},
	}
}

type session struct {
	sync.Mutex
	voters      map[interface{}]chan *event.Payload
	controllers map[interface{}]chan *event.Payload
}

func newSession() *session {
	return &session{
		voters:      map[interface{}]chan *event.Payload{},
		controllers: map[interface{}]chan *event.Payload{},
	}
}

func (i *Internal) Emit(sessionID string, r event.Role, t event.Type, body interface{}) {
	s, ok := i.sessions[sessionID]
	if !ok {
		return
	}

	s.Lock()
	defer s.Unlock()

	switch r {
	case event.Voter:
		for _, c := range s.voters {
			c <- &event.Payload{
				Kind: t,
				Data: body,
			}
		}
	case event.Controller:
		for _, c := range s.controllers {
			c <- &event.Payload{
				Kind: t,
				Data: body,
			}
		}
	}
}

func (i *Internal) Subscribe(sessionID string, r event.Role, ws interface{}) (chan *event.Payload, error) {
	log.Printf("subscribe %q", sessionID)
	c := make(chan *event.Payload)

	var s *session

	s, ok := i.sessions[sessionID]
	if !ok {
		i.Lock()
		defer i.Unlock()

		s = newSession()
		i.sessions[sessionID] = s
	}

	s.Lock()
	defer s.Unlock()

	switch r {
	case event.Voter:
		s.voters[ws] = c
	case event.Controller:
		s.controllers[ws] = c
	}

	return c, nil
}

func (i *Internal) Unsubscribe(sessionID string, ws interface{}) error {
	log.Printf("unsubscribe %q", sessionID)
	s, ok := i.sessions[sessionID]
	if !ok {
		return errors.New("no session found")
	}

	s.Lock()
	defer s.Unlock()

	if c, ok := s.voters[ws]; ok {
		close(c)
		delete(s.voters, ws)
	}

	if c, ok := s.controllers[ws]; ok {
		close(c)
		delete(s.controllers, ws)
	}

	if len(s.voters)+len(s.controllers) == 0 {
		delete(i.sessions, sessionID)
	}

	return nil
}
