package store

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/akarasz/pajthy-backend/domain"
)

var (
	ErrAlreadyExists = errors.New("session already exists")
	ErrNotExists     = errors.New("session not exists")
	ErrLocking       = errors.New("locking error")

	errVersionMismatch = errors.New("version mismatch")
)

type Store interface {
	Create(id string, created *domain.Session) error
	Update(id string, updated *Session) error
	Load(id string) (*Session, error)
}

type Session struct {
	Data    *domain.Session
	Version uuid.UUID
}

func OptimisticLocking(f func() error) error {
	for retry := 0; retry < 5; retry++ {
		err := f()
		if err == errVersionMismatch {
			time.Sleep(20 * time.Millisecond)
			continue
		}

		return err
	}

	return ErrLocking
}
