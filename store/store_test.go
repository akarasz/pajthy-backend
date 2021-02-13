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
	s := store.NewInMemory()

	// loading a non-existent session should return an error
	_, err := s.Load("id")
	t.Equal(err, store.ErrNotExists)

	created := domain.NewSession()
	t.Require().NoError(s.Create("id", created))

	// loading should return the session
	if got, err := s.Load("id"); t.NoError(err) {
		t.Exactly(created, got)
	}
}

func (t *Suite) TestCreate() {
	s := store.NewInMemory()

	// created session should be returned by load
	want := domain.NewSession()
	t.NoError(s.Create("id", want))

	got, err := s.Load("id")
	t.Require().NoError(err)
	t.Exactly(want, got)

	// creating with an existing id should return an error
	t.Equal(s.Create("id", want), store.ErrAlreadyExists)
}

func (t *Suite) TestUpdate() {
	s := store.NewInMemory()

	// update non existing id should return error
	t.Equal(store.ErrNotExists, s.Update("id", domain.NewSession()))

	created := domain.NewSession()
	t.Require().NoError(s.Create("id", created))

	// updating with different version should return error
	wrongVersion := domain.NewSession()
	wrongVersion.Version = uuid.Must(uuid.NewRandom())

	t.Error(s.Update("id", wrongVersion))

	// loading after updating right should return the same as updated
	updated := domain.NewSession()
	updated.Version = created.Version

	t.Require().NoError(s.Update("id", updated))

	loaded, err := s.Load("id")
	t.Require().NoError(err)

	t.Exactly(&loaded, &updated)
}
