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
	t.Require().NoError(s.Save("loadID", created))

	// loading should return the session
	if got, err := s.Load("loadID"); t.NoError(err) {
		t.Exactly(created, got.Data)
	}
}

func (t *Suite) TestSave() {
	s := t.Subject

	// save to non-existing id with version should fail
	t.Exactly(
		store.ErrVersionMismatch,
		s.Save("saveID", domain.NewSession(), uuid.Must(uuid.NewRandom())))

	// save with non-existing id and without version should be ok
	created := domain.NewSession()
	t.NoError(s.Save("saveID", created))

	// loading a saved item should return that item
	saved, err := s.Load("saveID")
	t.Require().NoError(err)
	t.Exactly(created, saved.Data)

	// save with existing id and wrong version should fail
	t.Exactly(
		store.ErrVersionMismatch,
		s.Save("saveID", domain.NewSession(), uuid.Must(uuid.NewRandom())))

	// saving with multiple versions should fail
	t.Exactly(
		store.ErrVersionMismatch,
		s.Save("saveID", domain.NewSession(), saved.Version, uuid.Must(uuid.NewRandom())))

	// save with existing id and right version should be ok
	modified := domain.NewSession()
	t.NoError(s.Save("saveID", modified, saved.Version))

	// loading after updating an item should return the updated item
	saved, err = s.Load("saveID")
	t.Require().NoError(err)
	t.Exactly(modified, saved.Data)
}
