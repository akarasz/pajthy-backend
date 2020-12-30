package store_test

import (
	"testing"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
)

func TestLoad(t *testing.T) {
	s := store.New()

	// loading a non-existent session should return an error
	_, err := s.Load("id")
	if want := store.ErrNotExists; err != want {
		t.Errorf("wrong error when loading non-existent session: %v", err)
	}

	// loading a non-locked id should return an error
	created := domain.NewSession()
	if err := s.Create("id", created); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	_, err = s.Load("id")
	if want := store.ErrNotLocked; err != want {
		t.Errorf("wrong error when loading non-locked session: %v", err)
	}

	// loading should return the session
	lock(t, s, "id")
	got, err := s.Load("id")
	if err != nil {
		t.Fatalf("unexpected error when loading: %v", err)
	}
	if want := created; got != want {
		t.Errorf("wrong session returned. got %v want %v", got, want)
	}
	unlock(t, s, "id")
}

func TestCreate(t *testing.T) {
	s := store.New()

	// created session should be returned by load
	want := domain.NewSession()
	if err := s.Create("id", want); err != nil {
		t.Errorf("failed to create session: %v", err)
	}

	lock(t, s, "id")
	got, err := s.Load("id")
	if err != nil {
		t.Fatalf("failed to load session just created: %v", err)
	}
	if got != want {
		t.Errorf("not the same session loaded that go created. got %v want %v", got, want)
	}
	unlock(t, s, "id")

	// creating with an existing id should return an error
	if err := s.Create("id", want); err != store.ErrAlreadyExists {
		t.Errorf("wrong error when creating with already existing id: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	s := store.New()

	// update non existing id should return error
	if err := s.Update("id", domain.NewSession()); err != store.ErrNotExists {
		t.Errorf("wrong error when trying to update non-existent session: %v", err)
	}

	// update existing without locking should return error
	created := domain.NewSession()
	if err := s.Create("id", created); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	if err := s.Update("id", domain.NewSession()); err != store.ErrNotLocked {
		t.Errorf("wrong error when updating without locking: %v", err)
	}

	// loading after updating right should return the same as updated
	lock(t, s, "id")
	updated := domain.NewSession()
	if err := s.Update("id", updated); err != nil {
		t.Fatalf("unexpected error while updating: %v", err)
	}
	loaded, err := s.Load("id")
	if err != nil {
		t.Fatalf("unable to load just updated session: %v", err)
	}
	if loaded != updated {
		t.Errorf("not the updated session was loaded. got %v want %v", loaded, updated)
	}
	unlock(t, s, "id")
}

func lock(t *testing.T, s *store.Store, id string) {
	if err := s.Lock(id); err != nil {
		t.Fatalf("unexpected error while locking: %v", err)
	}
}

func unlock(t *testing.T, s *store.Store, id string) {
	if err := s.Unlock(id); err != nil {
		t.Fatalf("unexpected error while unlocking: %v", err)
	}
}
