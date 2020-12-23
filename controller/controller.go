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
	s.Choices = choices

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

	return s.Choices, nil
}

func Vote(id string, v *domain.Vote) error {
	s, err := store.Load(id)
	if err != nil {
		return err
	}

	if !s.Open {
		return ErrSessionIsClosed
	}

	exists := false
	for _, p := range s.Participants {
		if p == v.Participant {
			exists = true
			break
		}
	}
	if !exists {
		return ErrVoterNotExists
	}

	exists = false
	for _, c := range s.Choices {
		if c == v.Choice {
			exists = true
			break
		}
	}
	if !exists {
		return ErrInvalidChoice
	}

	s.Votes[v.Participant] = v.Choice

	if len(s.Votes) == len(s.Participants) {
		s.Open = false
		event.EmitVoteDisabled(id)
	}

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	event.EmitVote(id, s.Votes)

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

	s.Open = true
	s.Votes = map[string]string{}

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	event.EmitVoteEnabled(id)

	return nil
}

func StopVote(id string) error {
	s, err := store.Load(id)
	if err != nil {
		return err
	}

	s.Open = false

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	event.EmitVoteDisabled(id)

	return nil
}

func ResetVote(id string) error {
	s, err := store.Load(id)
	if err != nil {
		return err
	}

	s.Open = false
	s.Votes = map[string]string{}

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	event.EmitReset(id)
	event.EmitVote(id, s.Votes)

	return nil
}

func KickParticipant(id string, name string) error {
	s, err := store.Load(id)
	if err != nil {
		return err
	}

	idx := -1
	for i, p := range s.Participants {
		if p == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrVoterNotExists
	}

	s.Participants = append(s.Participants[:idx], s.Participants[idx+1:]...)

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	event.EmitParticipantsChange(id, s.Participants)

	return nil
}

func Join(id string, name string) error {
	s, err := store.Load(id)
	if err != nil {
		return err
	}

	for _, p := range s.Participants {
		if p == name {
			return ErrVoterAlreadyJoined
		}
	}

	s.Participants = append(s.Participants, name)

	err = store.Save(id, s)
	if err != nil {
		return err
	}

	event.EmitParticipantsChange(id, s.Participants)

	return nil
}
