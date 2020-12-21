package event

import (
	"errors"
	"log"
	"sync"

	"github.com/akarasz/pajthy-backend/domain"
)

type role int

const (
	engineer role = iota
	scrumMaster
)

type event string

const (
	eventEnabled            = "enabled"
	eventDisabled           = "disabled"
	eventParticipantsChange = "participants-change"
	eventVote               = "vote"
	eventDone               = "done"
)

type message struct {
	Kind event
	Data interface{}
}

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

func emit(sessionID string, r role, body *message) {
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
	log.Printf("emit enabled %q", sessionID)
	m := &message{
		Kind: eventEnabled,
	}
	emit(sessionID, engineer, m)
	emit(sessionID, scrumMaster, m)
}

func EmitVoteDisabled(sessionID string) {
	log.Printf("emit disabled %q", sessionID)
	m := &message{
		Kind: eventDisabled,
	}
	emit(sessionID, engineer, m)
	emit(sessionID, scrumMaster, m)
}

func EmitParticipantsChange(sessionID string, participants []string) {
	log.Printf("emit participants change %q %q", sessionID, participants)
	emit(sessionID, scrumMaster, &message{
		Kind: eventParticipantsChange,
		Data: participants,
	})
}

func EmitVote(sessionID string, v *domain.Vote) {
	log.Printf("emit vote %q %q", sessionID, v)
	emit(sessionID, scrumMaster, &message{
		Kind: eventVote,
		Data: v,
	})
}

func EmitDone(sessionID string, votes map[string]string) {
	log.Printf("emit done %q %q", sessionID, votes)
	EmitVoteDisabled(sessionID)
	emit(sessionID, scrumMaster, &message{
		Kind: eventDone,
		Data: votes,
	})
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
	log.Printf("subscribe engineer %q", sessionID)
	return subscribe(sessionID, engineer, ws)
}

func SubscribeScrumMaster(sessionID string, ws interface{}) (chan interface{}, error) {
	log.Printf("subscribe scrum master %q", sessionID)
	return subscribe(sessionID, scrumMaster, ws)
}

func Unsubscribe(sessionID string, ws interface{}) error {
	log.Printf("unsubscribe %q", sessionID)
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