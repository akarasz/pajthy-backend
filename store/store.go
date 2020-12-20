package storage

import (
	"errors"

	"github.com/akarasz/pajthy-backend/domain"
)

var (
	ErrNotExists = errors.New("not exists")
)

var repository map[string]*domain.Session

func init() {
	repository = map[string]*domain.Session{}
}

func Save(id string, s *domain.Session) error {
	repository[id] = s

	return nil
}

func Load(id string) (*domain.Session, error) {
	s, ok := repository[id]
	if !ok {
		return s, ErrNotExists
	}

	return s, nil
}