package store

import (
	"errors"

	"github.com/akarasz/pajthy-backend/domain"
)

var (
	ErrNotExists = errors.New("session not exists")
)

type Store interface {
	Save(id string, s *domain.Session) error
	Load(id string) (*domain.Session, error)
}

func NewInternal() Store {
	return &Internal{
		repo: map[string]*domain.Session{},
	}
}

type Internal struct {
	repo map[string]*domain.Session
}

func (i *Internal) Save(id string, s *domain.Session) error {
	i.repo[id] = s

	return nil
}

func (i *Internal) Load(id string) (*domain.Session, error) {
	s, ok := i.repo[id]
	if !ok {
		return s, ErrNotExists
	}

	return s, nil
}
