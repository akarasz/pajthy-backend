package store

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/akarasz/pajthy-backend/domain"
)

var (
	ErrNotExists       = errors.New("session not exists")
	ErrVersionMismatch = errors.New("version mismatch")
)

type Store interface {
	Load(id string) (*Session, error)
	Save(id string, item *domain.Session, version ...uuid.UUID) error
}

type Session struct {
	Data    *domain.Session
	Version uuid.UUID
}

func WithNewVersion(data *domain.Session) *Session {
	return &Session{
		Data:    data,
		Version: uuid.Must(uuid.NewRandom()),
	}
}

func ReadModifyWrite(id string, s Store, modify func(*domain.Session) (*domain.Session, error)) (*domain.Session, error) {
	for retry := 0; retry < 5; retry++ {
		loaded, err := s.Load(id)
		if err != nil {
			return nil, err
		}

		modified, err := modify(loaded.Data)
		if err != nil {
			return nil, err
		}

		if err = s.Save(id, modified, loaded.Version); err != nil {
			if err == ErrVersionMismatch {
				time.Sleep(20 * time.Millisecond)
				continue
			}

			return nil, err
		}

		return modified, nil
	}

	return nil, ErrVersionMismatch
}
