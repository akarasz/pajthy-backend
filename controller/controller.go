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

type Controller struct {
	store store.Store
	event event.Event
}

func New(s store.Store, e event.Event) *Controller {
	return &Controller{
		store: s,
		event: e,
	}
}

func (c *Controller) CreateSession(choices []string) (string, error) {
	id := generateID()

	s := domain.NewSession()
	s.Choices = choices

	if err := c.store.Save(id, s); err != nil {
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

type choicesResponse struct {
	Choices []string
	Open    bool
}

func (c *Controller) Choices(id string) (*choicesResponse, error) {
	s, err := c.store.Load(id)
	if err != nil {
		return nil, err
	}

	return &choicesResponse{
		Choices: s.Choices,
		Open:    s.Open,
	}, nil
}

func (c *Controller) Vote(id string, v *domain.Vote) error {
	s, err := c.store.Load(id)
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
		c.emitVoteDisabled(id)
	}

	err = c.store.Save(id, s)
	if err != nil {
		return err
	}

	c.emitVote(id, s.Votes)

	return nil
}

func (c *Controller) GetSession(id string) (*domain.Session, error) {
	s, err := c.store.Load(id)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Controller) StartVote(id string) error {
	s, err := c.store.Load(id)
	if err != nil {
		return err
	}

	s.Open = true
	s.Votes = map[string]string{}

	err = c.store.Save(id, s)
	if err != nil {
		return err
	}

	c.emitVoteEnabled(id)

	return nil
}

func (c *Controller) StopVote(id string) error {
	s, err := c.store.Load(id)
	if err != nil {
		return err
	}

	s.Open = false

	err = c.store.Save(id, s)
	if err != nil {
		return err
	}

	c.emitVoteDisabled(id)

	return nil
}

func (c *Controller) ResetVote(id string) error {
	s, err := c.store.Load(id)
	if err != nil {
		return err
	}

	s.Open = false
	s.Votes = map[string]string{}

	err = c.store.Save(id, s)
	if err != nil {
		return err
	}

	c.emitReset(id)
	c.emitVote(id, s.Votes)

	return nil
}

func (c *Controller) KickParticipant(id string, name string) error {
	s, err := c.store.Load(id)
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

	err = c.store.Save(id, s)
	if err != nil {
		return err
	}

	c.emitParticipantsChange(id, s.Participants)

	return nil
}

func (c *Controller) Join(id string, name string) error {
	s, err := c.store.Load(id)
	if err != nil {
		return err
	}

	for _, p := range s.Participants {
		if p == name {
			return ErrVoterAlreadyJoined
		}
	}

	s.Participants = append(s.Participants, name)

	err = c.store.Save(id, s)
	if err != nil {
		return err
	}

	c.emitParticipantsChange(id, s.Participants)

	return nil
}

type openChangedData struct {
	Open bool
}

type participantsChangedData struct {
	Participants []string
}

type votesChangedData struct {
	Votes map[string]string
}

func (c *Controller) emitVoteEnabled(id string) {
	m := &openChangedData{Open: true}
	c.event.Emit(id, event.Voter, event.Enabled, m)
	c.event.Emit(id, event.Controller, event.Enabled, m)
}

func (c *Controller) emitVoteDisabled(id string) {
	m := &openChangedData{Open: false}
	c.event.Emit(id, event.Voter, event.Disabled, m)
	c.event.Emit(id, event.Controller, event.Disabled, m)
}

func (c *Controller) emitReset(id string) {
	m := &openChangedData{Open: false}
	c.event.Emit(id, event.Voter, event.Reset, m)
	c.event.Emit(id, event.Controller, event.Reset, m)
}

func (c *Controller) emitParticipantsChange(id string, participants []string) {
	c.event.Emit(
		id,
		event.Controller,
		event.ParticipantsChange,
		&participantsChangedData{Participants: participants})
}

func (c *Controller) emitVote(id string, votes map[string]string) {
	c.event.Emit(id, event.Controller, event.Vote, &votesChangedData{Votes: votes})
}
