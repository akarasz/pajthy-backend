package store_test

import (
	"testing"

	"github.com/google/uuid"
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

	created := domain.NewSession()
	require.NoError(t, s.Create("id", created))

	// loading should return the session
	if got, err := s.Load("id"); assert.NoError(t, err) {
		assert.Exactly(t, created, got)
	}
}

func TestCreate(t *testing.T) {
	s := store.New()

	// created session should be returned by load
	want := domain.NewSession()
	assert.NoError(t, s.Create("id", want))

	got, err := s.Load("id")
	require.NoError(t, err)
	assert.Exactly(t, want, got)

	// creating with an existing id should return an error
	assert.Equal(t, s.Create("id", want), store.ErrAlreadyExists)
}

func TestUpdate(t *testing.T) {
	s := store.New()

	// update non existing id should return error
	assert.Equal(t, store.ErrNotExists, s.Update("id", domain.NewSession()))

	created := domain.NewSession()
	require.NoError(t, s.Create("id", created))

	// updating with different version should return error
	wrongVersion := domain.NewSession()
	wrongVersion.Version = uuid.Must(uuid.NewRandom())

	assert.Error(t, s.Update("id", wrongVersion))

	// loading after updating right should return the same as updated
	updated := domain.NewSession()
	updated.Version = created.Version

	require.NoError(t, s.Update("id", updated))

	loaded, err := s.Load("id")
	require.NoError(t, err)

	assert.Exactly(t, &loaded, &updated)
}
