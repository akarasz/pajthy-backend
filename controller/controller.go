package controller

import (
	"errors"
	"math/rand"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/event"
	"github.com/akarasz/pajthy-backend/store"
)

var (
	ErrSessionIsClosed    = errors.New("session is closed")
	ErrVoterAlreadyJoined = errors.New("voter already joined")
	ErrVoterNotExists     = errors.New("voter not exists")
	ErrInvalidChoice      = errors.New("invalid choice")
)

func CreateSession(choices []string) (string, error) {
	id := generateID()

	s := domain.NewSession()
	for _, c := range choices {
		s.Choices[c] = struct{}{}
	}

	if err := store.Save(id, s); err != nil {
		return id, err
	}

	return id, nil
}

func generateID() string {
	const (
		idCharset = "abcdefghijklmnopqrstvwxyz0123456789"
		length    = 5
	)

	b := make([]byte, length)
	for i := range b {
		b[i] = idCharset[rand.Intn(len(idCharset))]
	}
	return string(b)
}

func Choices(id string) ([]string, error) {
	s, err := store.Load(id)
	if err != nil {
		return []string{}, err
	}

	res := []string{}
	for k := range s.Choices {
		res = append(res, k)
	}

	return res, nil
}

func Vote(id string, v *domain.Vote) error {
	s, err := store.Load(id)
	if err != nil {
		return err
	}

	if !s.Open {
		return ErrSessionIsClosed
	}

	if _, exists := s.Participants[v.Participant]; !exists {
		return ErrVoterNotExists
	}

	if _, exists := s.Choices[v.Choice]; !exists {
		return ErrInvalidChoice
	}

	alreadyVoted := false
	for _, existing := range s.Votes {
		if v.Participant == existing.Participant {
			alreadyVoted = true
			existing.Choice = v.Choice
		}
	}
	if !alreadyVoted {
		s.Votes = append(s.Votes, v)
	}

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	if len(s.Votes) == len(s.Participants) {
		event.EmitVoteDisabled(id)
	}

	return nil
}

func GetSession(id string) (*domain.Session, error) {
	s, err := store.Load(id)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func StartVote(id string) error {
	s, err := store.Load(id)
	if err != nil {
		return err
	}

	s.Votes = []*domain.Vote{}
	s.Open = true

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	event.EmitVoteEnabled(id)

	return nil
}

func ResetVote(id string) error {
	s, err := store.Load(id)
	if err != nil {
		return err
	}

	s.Open = false
	s.Votes = []*domain.Vote{}

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	event.EmitVoteDisabled(id)

	return nil
}

func KickParticipant(id string, name string) error {
	s, err := store.Load(id)
	if err != nil {
		return err
	}

	if _, exists := s.Participants[name]; exists {
		delete(s.Participants, name)
	}

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	return nil
}

func Join(id string, name string) error {
	s, err := store.Load(id)
	if err != nil {
		return err
	}

	if _, exists := s.Participants[name]; exists {
		return ErrVoterAlreadyJoined
	}

	s.Participants[name] = struct{}{}

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	return nil
}
