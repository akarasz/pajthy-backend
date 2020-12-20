package event

import (
	"errors"
	"sync"
)

type role int

const (
	engineer role = iota
	scrumMaster
)

type session struct {
	sync.Mutex
	engineers    map[interface{}]chan interface{}
	scrumMasters map[interface{}]chan interface{}
}

func newSession() *session {
	return &session{
		engineers:    map[interface{}]chan interface{}{},
		scrumMasters: map[interface{}]chan interface{}{},
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

func emit(sessionID string, r role, body interface{}) {
	s, ok := repository.sessions[sessionID]
	if !ok {
		return
	}

	s.Lock()
	defer s.Unlock()

	switch r {
	case engineer:
		for _, c := range s.engineers {
			c <- body
		}
	case scrumMaster:
		for _, c := range s.scrumMasters {
			c <- body
		}
	}
}

func EmitVoteEnabled(sessionID string) {
	emit(sessionID, engineer, "enabled")
}

func EmitVoteDisabled(sessionID string) {
	emit(sessionID, engineer, "disabled")
}

func subscribe(sessionID string, r role, ws interface{}) (chan interface{}, error) {
	c := make(chan interface{})

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
	case engineer:
		s.engineers[ws] = c
	case scrumMaster:
		s.scrumMasters[ws] = c
	}

	return c, nil
}

func SubscribeEngineer(sessionID string, ws interface{}) (chan interface{}, error) {
	return subscribe(sessionID, engineer, ws)
}

func SubscribeScrumMaster(sessionID string, ws interface{}) (chan interface{}, error) {
	return subscribe(sessionID, scrumMaster, ws)
}

func Unsubscribe(sessionID string, ws interface{}) error {
	s, ok := repository.sessions[sessionID]
	if !ok {
		return errors.New("no session found")
	}

	s.Lock()
	defer s.Unlock()

	if c, ok := s.engineers[ws]; ok {
		close(c)
		delete(s.engineers, ws)
	}

	if c, ok := s.scrumMasters[ws]; ok {
		close(c)
		delete(s.scrumMasters, ws)
	}

	if len(s.engineers)+len(s.scrumMasters) == 0 {
		delete(repository.sessions, sessionID)
	}

	return nil
}
