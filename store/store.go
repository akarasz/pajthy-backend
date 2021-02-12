package store

import (
	"errors"
	"sync"
	"time"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/google/uuid"
)

var (
	ErrAlreadyExists = errors.New("session already exists")
	ErrNotExists     = errors.New("session not exists")
	ErrLocking       = errors.New("locking error")

	errVersionMismatch = errors.New("version mismatch")
)

type Store struct {
	repo map[string]*domain.Session

	sync.RWMutex
}

func New() *Store {
	return &Store{
		repo: map[string]*domain.Session{},
	}
}

func (s *Store) Create(id string, created *domain.Session) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.repo[id]
	if ok {
		return ErrAlreadyExists
	}

	s.repo[id] = created

	return nil
}

func (s *Store) Update(id string, updated *domain.Session) error {
	s.Lock()
	defer s.Unlock()

	current, ok := s.repo[id]
	if !ok {
		return ErrNotExists
	}

	if current.Version != updated.Version {
		return errVersionMismatch
	}

	updated.Version = uuid.Must(uuid.NewRandom())
	s.repo[id] = updated
	return nil
}

func (s *Store) Load(id string) (*domain.Session, error) {
	s.RLock()
	defer s.RUnlock()

	saved, ok := s.repo[id]
	if !ok {
		return nil, ErrNotExists
	}

	return saved, nil
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
