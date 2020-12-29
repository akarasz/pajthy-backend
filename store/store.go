package store

import (
	"errors"

	"github.com/akarasz/pajthy-backend/domain"
)

var (
	ErrAlreadyExists = errors.New("session already exists")
	ErrNotExists     = errors.New("session not exists")
	ErrNotLocked     = errors.New("key is not locked")
)

type Store interface {
	Lock(id string) error
	Unlock(id string) error

	Create(id string, s *domain.Session) error
	Update(id string, s *domain.Session) error
	Load(id string) (*domain.Session, error)

	LockAndLoad(id string) (*domain.Session, error)
}
