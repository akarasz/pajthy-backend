package controller

import (
	"github.com/akarasz/pajthy-backend/domain"
)

func CreateSession() (string, error) {
	return "abcde", nil
}

func Choices(id string) ([]string, error) {
	return []string{"option #1", "option #2", "option #3"}, nil
}

func Vote(id string, v *domain.Vote) error {
	return nil
}

func GetSession(id string) (*domain.Session, error) {
	return &domain.Session{
		Choices:      []string{},
		Participants: []string{},
		Votes:        []domain.Vote{},
		Open:         false,
	}, nil
}

func StartVote(id string) error {
	return nil
}

func ResetVote(id string) error {
	return nil
}

func KickParticipant(id string, name string) error {
	return nil
}

func Join(id string, name string) error {
	return nil
}
