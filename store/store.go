package store

import (
	"errors"
	"sync"

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

func NewInternal() Store {
	return &Internal{
		repo:       map[string]*sessionWithLock{},
		createLock: &sync.Mutex{},
	}
}

func (i *Internal) Lock(id string) error {
	s, ok := i.repo[id]
	if !ok {
		return ErrNotExists
	}

	s.Lock()
	s.locked = true
	return nil
}

func (i *Internal) Unlock(id string) error {
	s, ok := i.repo[id]
	if !ok {
		return ErrNotExists
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
		return ErrAlreadyExists
	}

	i.repo[id] = newSessionWithLock(s)

	return nil
}

func (i *Internal) Update(id string, s *domain.Session) error {
	swl, ok := i.repo[id]
	if !ok {
		return ErrNotExists
	}

	if !swl.locked {
		return ErrNotLocked
	}

	swl.session = s

	return nil
}

func (i *Internal) Load(id string) (*domain.Session, error) {
	swl, ok := i.repo[id]
	if !ok {
		return nil, ErrNotExists
	}

	if !swl.locked {
		return nil, ErrNotLocked
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
