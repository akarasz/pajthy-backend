package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
)

func TestLoad(t *testing.T) {
	s := store.New()

	// loading a non-existent session should return an error
	_, err := s.Load("id")
	assert.Equal(t, err, store.ErrNotExists)

	// loading a non-locked id should return an error
	created := domain.NewSession()
	require.NoError(t, s.Create("id", created))

	_, err = s.Load("id")
	assert.Equal(t, err, store.ErrNotLocked)

	// loading should return the session
	lock(t, s, "id")
	if got, err := s.Load("id"); assert.NoError(t, err) {
		assert.Exactly(t, created, got)
	}
	unlock(t, s, "id")
}

func TestCreate(t *testing.T) {
	s := store.New()

	// created session should be returned by load
	want := domain.NewSession()
	assert.NoError(t, s.Create("id", want))

	lock(t, s, "id")
	got, err := s.Load("id")
	require.NoError(t, err)
	assert.Exactly(t, want, got)
	unlock(t, s, "id")

	// creating with an existing id should return an error
	assert.Equal(t, s.Create("id", want), store.ErrAlreadyExists)
}

func TestUpdate(t *testing.T) {
	s := store.New()

	// update non existing id should return error
	assert.Equal(t, store.ErrNotExists, s.Update("id", domain.NewSession()))

	// update existing without locking should return error
	created := domain.NewSession()
	require.NoError(t, s.Create("id", created))

	assert.Equal(t, store.ErrNotLocked, s.Update("id", domain.NewSession()))

	// loading after updating right should return the same as updated
	lock(t, s, "id")

	updated := domain.NewSession()
	require.NoError(t, s.Update("id", updated))

	loaded, err := s.Load("id")
	require.NoError(t, err)

	assert.Exactly(t, loaded, updated)

	unlock(t, s, "id")
}

func lock(t *testing.T, s *store.Store, id string) {
	require.NoError(t, s.Lock(id))
}

func unlock(t *testing.T, s *store.Store, id string) {
	require.NoError(t, s.Unlock(id))
}
