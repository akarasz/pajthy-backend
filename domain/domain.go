package domain

import (
	"github.com/google/uuid"
)

type Vote struct {
	Participant string
	Choice      string
}

type Session struct {
	Choices      []string
	Participants []string
	Votes        map[string]string
	Open         bool
	Version      uuid.UUID `json:"-"`
}

func NewSession() *Session {
	return &Session{
		Choices:      []string{},
		Participants: []string{},
		Votes:        map[string]string{},
		Open:         false,
		Version:      uuid.Must(uuid.NewRandom()),
	}
}
