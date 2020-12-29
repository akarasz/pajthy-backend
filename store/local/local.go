package local

import (
	"sync"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
)

type sessionWithLock struct {
	sync.Mutex

	session *domain.Session
	locked  bool
}

func newSessionWithLock(s *domain.Session) *sessionWithLock {
	return &sessionWithLock{
		session: s,
	}
}

type Internal struct {
	repo       map[string]*sessionWithLock
	createLock *sync.Mutex
}

func New() *Internal {
	return &Internal{
		repo:       map[string]*sessionWithLock{},
		createLock: &sync.Mutex{},
	}
}

func (i *Internal) Lock(id string) error {
	s, ok := i.repo[id]
	if !ok {
		return store.ErrNotExists
	}

	s.Lock()
	s.locked = true
	return nil
}

func (i *Internal) Unlock(id string) error {
	s, ok := i.repo[id]
	if !ok {
		return store.ErrNotExists
	}

	s.Unlock()
	s.locked = false
	return nil
}

func (i *Internal) Create(id string, s *domain.Session) error {
	i.createLock.Lock()
	defer i.createLock.Unlock()

	_, ok := i.repo[id]
	if ok {
		return store.ErrAlreadyExists
	}

	i.repo[id] = newSessionWithLock(s)

	return nil
}

func (i *Internal) Update(id string, s *domain.Session) error {
	swl, ok := i.repo[id]
	if !ok {
		return store.ErrNotExists
	}

	if !swl.locked {
		return store.ErrNotLocked
	}

	swl.session = s

	return nil
}

func (i *Internal) Load(id string) (*domain.Session, error) {
	swl, ok := i.repo[id]
	if !ok {
		return nil, store.ErrNotExists
	}

	if !swl.locked {
		return nil, store.ErrNotLocked
	}

	return swl.session, nil
}

func (i *Internal) LockAndLoad(id string) (*domain.Session, error) {
	if err := i.Lock(id); err != nil {
		return nil, err
	}

	res, err := i.Load(id)
	if err != nil {
		i.Unlock(id)
	}
	return res, err
}
