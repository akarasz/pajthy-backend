package store_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
)

type Suite struct {
	suite.Suite

	Subject store.Store
}

func (t *Suite) TestLoad() {
	s := t.Subject

	// loading a non-existent session should return an error
	_, err := s.Load("loadID")
	t.Equal(store.ErrNotExists, err)

	created := domain.NewSession()
	t.Require().NoError(s.Create("loadID", created))

	// loading should return the session
	if got, err := s.Load("loadID"); t.NoError(err) {
		t.Exactly(created, got.Data)
	}
}

func (t *Suite) TestCreate() {
	s := t.Subject

	// created session should be returned by load
	want := domain.NewSession()
	t.NoError(s.Create("createID", want))

	got, err := s.Load("createID")
	t.Require().NoError(err)
	t.Exactly(want, got.Data)

	// creating with an existing id should return an error
	t.Equal(store.ErrAlreadyExists, s.Create("createID", want))
}

func (t *Suite) TestUpdate() {
	s := t.Subject

	// update non existing id should return error
	t.Equal(store.ErrNotExists, s.Update("updateID", &store.Session{}))

	created := domain.NewSession()
	t.Require().NoError(s.Create("updateID", created))

	// updating with different version should return error
	wrongVersion := &store.Session{
		Data:    created,
		Version: uuid.Must(uuid.NewRandom()),
	}

	t.Error(s.Update("updateID", wrongVersion))

	// loading right after updating should return the same as updated
	stored, err := s.Load("updateID")
	t.Require().NoError(err)

	updated := &store.Session{
		Data:    domain.NewSession(),
		Version: stored.Version,
	}

	t.Require().NoError(s.Update("updateID", updated))

	loaded, err := s.Load("updateID")
	t.Require().NoError(err)

	t.Exactly(updated.Data, loaded.Data)
}
