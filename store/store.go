package store

import (
	"errors"
	"sync"

	"github.com/akarasz/pajthy-backend/domain"
)

var (
	ErrAlreadyExists = errors.New("session already exists")
	ErrNotExists     = errors.New("session not exists")
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

type Store struct {
	repo       map[string]*sessionWithLock
	createLock *sync.Mutex
}

func New() *Store {
	return &Store{
		repo:       map[string]*sessionWithLock{},
		createLock: &sync.Mutex{},
	}
}

func (s *Store) Lock(id string) error {
	swl, ok := s.repo[id]
	if !ok {
		return ErrNotExists
	}

	swl.Lock()
	swl.locked = true
	return nil
}

func (s *Store) Unlock(id string) error {
	swl, ok := s.repo[id]
	if !ok {
		return ErrNotExists
	}

	swl.Unlock()
	swl.locked = false
	return nil
}

func (s *Store) Create(id string, sess *domain.Session) error {
	s.createLock.Lock()
	defer s.createLock.Unlock()

	_, ok := s.repo[id]
	if ok {
		return ErrAlreadyExists
	}

	s.repo[id] = newSessionWithLock(sess)

	return nil
}

func (s *Store) Update(id string, sess *domain.Session) error {
	swl, ok := s.repo[id]
	if !ok {
		return ErrNotExists
	}

	swl.session = sess

	return nil
}

func (s *Store) Load(id string) (*domain.Session, error) {
	swl, ok := s.repo[id]
	if !ok {
		return nil, ErrNotExists
	}

	return swl.session, nil
}

func (s *Store) LockAndLoad(id string) (*domain.Session, error) {
	if err := s.Lock(id); err != nil {
		return nil, err
	}

	res, err := s.Load(id)
	if err != nil {
		s.Unlock(id)
	}
	return res, err
}
