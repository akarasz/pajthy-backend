package store_test

import (
	"testing"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
)

func TestLoad(t *testing.T) {
	_, err := store.Load("load-id")

	if want := store.ErrNotExists; err != want {
		t.Errorf("wrong error on loading not existing item: %q", err)
	}

	want := domain.NewSession()
	err = store.Save("load-id", want)
	if err != nil {
		t.Fatalf("error saving item: %q", err)
	}

	got, err := store.Load("load-id")
	if err != nil {
		t.Fatalf("loading returned an unexpected error: %q", err)
	}
	if got != want {
		t.Errorf("returned an unexpected result: %v, wanted: %v", got, want)
	}

	got, err = store.Load("load-id")
	if err != nil {
		t.Fatalf("loading returned an unexpected error: %q", err)
	}
	if got != want {
		t.Errorf("loading again returned an unexpected result: %v, wanted: %v", got, want)
	}
}

func TestSave(t *testing.T) {
	want := domain.NewSession()
	err := store.Save("save-id", want)
	if err != nil {
		t.Fatalf("error saving item: %q", err)
	}

	got, err := store.Load("save-id")
	if err != nil {
		t.Fatalf("loading returned an unexpected error: %q", err)
	}
	if got != want {
		t.Errorf("different item loaded than saved; loaded %v saved %v", got, want)
	}

	overrideWith := domain.NewSession()
	err = store.Save("save-id", overrideWith)
	if err != nil {
		t.Fatalf("error overriding item: %q", err)
	}

	got, err = store.Load("save-id")
	if err != nil {
		t.Fatalf("loading returned an unexpected error: %q", err)
	}
	if got != overrideWith {
		t.Errorf("different item loaded than saved after override; loaded %v saved %v", got, overrideWith)
	}
}