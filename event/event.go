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
	sync.Mutex
	sessions map[string]*session
}

func New() *Event {
	return &Event{
		sessions: map[string]*session{},
	}
}

type session struct {
	sync.Mutex
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
	s, ok := e.sessions[sessionID]
	if !ok {
		return
	}

	s.Lock()
	defer s.Unlock()

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
}

func (e *Event) Subscribe(sessionID string, r Role, ws interface{}) (chan *Payload, error) {
	log.Printf("subscribe %q", sessionID)
	c := make(chan *Payload)

	var s *session

	s, ok := e.sessions[sessionID]
	if !ok {
		e.Lock()
		defer e.Unlock()

		s = newSession()
		e.sessions[sessionID] = s
	}

	s.Lock()
	defer s.Unlock()

	switch r {
	case Voter:
		s.voters[ws] = c
	case Controller:
		s.controllers[ws] = c
	}

	return c, nil
}

func (e *Event) Unsubscribe(sessionID string, ws interface{}) error {
	log.Printf("unsubscribe %q", sessionID)
	s, ok := e.sessions[sessionID]
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
		delete(e.sessions, sessionID)
	}

	return nil
}
