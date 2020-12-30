package store_test

import (
	"testing"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
)

func TestLoad(t *testing.T) {
	s := store.New()

	_, err := s.Load("id")
	if want := store.ErrNotExists; err != want {
		t.Errorf("wrong error when loading non-existent session: %v", err)
	}

	created := domain.NewSession()
	if err := s.Create("id", created); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	_, err = s.Load("id")
	if want := store.ErrNotLocked; err != want {
		t.Errorf("wrong error when loading non-locked session: %v", err)
	}

	if err := s.Lock("id"); err != nil {
		t.Fatalf("failed to lock session: %v", err)
	}
	defer s.Unlock("id")

	got, err := s.Load("id")
	if err != nil {
		t.Fatalf("unexpected error when loading: %v", err)
	}
	if want := created; got != want {
		t.Errorf("wrong session returned. got %v want %v", got, want)
	}
}