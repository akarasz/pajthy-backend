package event

import (
	"errors"
	"log"
	"sync"
)

type Role int

const (
	Voter Role = iota
	Controller
)

type Type string

const (
	Enabled            = Type("enabled")
	Disabled           = Type("disabled")
	Reset              = Type("reset")
	ParticipantsChange = Type("participants-change")
	Vote               = Type("vote")
	Done               = Type("done")
)

type Payload struct {
	Kind Type
	Data interface{}
}

func NewPayload(kind Type, data interface{}) *Payload {
	return &Payload{
		Kind: kind,
		Data: data,
	}
}

type Event struct {
	sync.RWMutex
	sessions map[string]*session
}

func New() *Event {
	return &Event{
		sessions: map[string]*session{},
	}
}

type session struct {
	sync.RWMutex
	voters      map[interface{}]chan *Payload
	controllers map[interface{}]chan *Payload
}

func newSession() *session {
	return &session{
		voters:      map[interface{}]chan *Payload{},
		controllers: map[interface{}]chan *Payload{},
	}
}

func (e *Event) Emit(sessionID string, r Role, t Type, body interface{}) {
	e.RLock()
	s, exists := e.sessions[sessionID]
	e.RUnlock()

	if !exists {
		return
	}

	s.RLock()
	switch r {
	case Voter:
		for _, c := range s.voters {
			c <- NewPayload(t, body)
		}
	case Controller:
		for _, c := range s.controllers {
			c <- NewPayload(t, body)
		}
	}
	s.RUnlock()
}

func (e *Event) Subscribe(sessionID string, r Role, ws interface{}) (chan *Payload, error) {
	log.Printf("subscribe %q", sessionID)
	c := make(chan *Payload)

	var s *session

	e.RLock()
	s, exists := e.sessions[sessionID]
	e.RUnlock()
	if !exists {
		s = newSession()

		e.Lock()
		e.sessions[sessionID] = s
		e.Unlock()
	}

	s.Lock()
	switch r {
	case Voter:
		s.voters[ws] = c
	case Controller:
		s.controllers[ws] = c
	}
	s.Unlock()

	return c, nil
}

func (e *Event) Unsubscribe(sessionID string, ws interface{}) error {
	log.Printf("unsubscribe %q", sessionID)
	e.RLock()
	s, exists := e.sessions[sessionID]
	e.RUnlock()
	if !exists {
		return errors.New("no session found")
	}

	s.Lock()
	if c, ok := s.voters[ws]; ok {
		close(c)
		delete(s.voters, ws)
	}

	if c, ok := s.controllers[ws]; ok {
		close(c)
		delete(s.controllers, ws)
	}
	s.Unlock()

	s.RLock()
	e.Lock()
	if len(s.voters)+len(s.controllers) == 0 {
		delete(e.sessions, sessionID)
	}
	e.Unlock()
	s.RUnlock()

	return nil
}
