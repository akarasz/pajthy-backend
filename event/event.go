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
	Enabled            = "enabled"
	Disabled           = "disabled"
	Reset              = "reset"
	ParticipantsChange = "participants-change"
	Vote               = "vote"
	Done               = "done"
)

type Payload struct {
	Kind Type
	Data interface{}
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

type sessions struct {
	sync.Mutex
	sessions map[string]*session
}

var repository sessions

func init() {
	repository = sessions{
		sessions: map[string]*session{},
	}
}

func Emit(sessionID string, r Role, t Type, body interface{}) {
	s, ok := repository.sessions[sessionID]
	if !ok {
		return
	}

	s.Lock()
	defer s.Unlock()

	switch r {
	case Voter:
		for _, c := range s.voters {
			c <- &Payload{
				Kind: t,
				Data: body,
			}
		}
	case Controller:
		for _, c := range s.controllers {
			c <- &Payload{
				Kind: t,
				Data: body,
			}
		}
	}
}

func Subscribe(sessionID string, r Role, ws interface{}) (chan *Payload, error) {
	log.Printf("subscribe %q", sessionID)
	c := make(chan *Payload)

	var s *session

	s, ok := repository.sessions[sessionID]
	if !ok {
		repository.Lock()
		defer repository.Unlock()

		s = newSession()
		repository.sessions[sessionID] = s
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

func Unsubscribe(sessionID string, ws interface{}) error {
	log.Printf("unsubscribe %q", sessionID)
	s, ok := repository.sessions[sessionID]
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
		delete(repository.sessions, sessionID)
	}

	return nil
}
