package store

import (
	"sync"

	"github.com/google/uuid"

	"github.com/akarasz/pajthy-backend/domain"
)

type InMemory struct {
	repo map[string]*Session

	sync.RWMutex
}

func NewInMemory() *InMemory {
	return &InMemory{
		repo: map[string]*Session{},
	}
}

func (im *InMemory) Create(id string, created *domain.Session) error {
	im.Lock()
	defer im.Unlock()

	_, ok := im.repo[id]
	if ok {
		return ErrAlreadyExists
	}

	im.repo[id] = &Session{
		Data:    created,
		Version: uuid.Must(uuid.NewRandom()),
	}

	return nil
}

func (im *InMemory) Update(id string, updated *Session) error {
	im.Lock()
	defer im.Unlock()

	current, ok := im.repo[id]
	if !ok {
		return ErrNotExists
	}

	if current.Version != updated.Version {
		return errVersionMismatch
	}

	updated.Version = uuid.Must(uuid.NewRandom())
	im.repo[id] = updated
	return nil
}

func (im *InMemory) Load(id string) (*Session, error) {
	im.RLock()
	defer im.RUnlock()

	saved, ok := im.repo[id]
	if !ok {
		return nil, ErrNotExists
	}

	return saved, nil
}
