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

func (im *InMemory) Load(id string) (*Session, error) {
	im.RLock()
	defer im.RUnlock()

	saved, ok := im.repo[id]
	if !ok {
		return nil, ErrNotExists
	}

	return saved, nil
}

func (im *InMemory) Save(id string, item *domain.Session, version ...uuid.UUID) error {
	if len(version) > 1 {
		return ErrVersionMismatch
	}

	im.Lock()
	defer im.Unlock()

	current, exists := im.repo[id]
	if (exists && (len(version) == 0 || current.Version != version[0])) || (!exists && len(version) != 0) {
		return ErrVersionMismatch
	}

	im.repo[id] = WithNewVersion(item)
	return nil
}
