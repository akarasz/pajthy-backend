package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/handler"
	"github.com/akarasz/pajthy-backend/store"
)

func TestCreateSession(t *testing.T) {
	s := store.New()
	req, err := http.NewRequest("POST", "/", strings.NewReader(`["one", "two"]`))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := handler.NewRouter(s, nil)
	router.ServeHTTP(rr, req)

	// returns created
	if got, want := rr.Code, http.StatusCreated; got != want {
		t.Errorf("wrong return code. got %v want %v", got, want)
	}

	// id in location header where the session is saved in the store
	location, exists := rr.HeaderMap["Location"]
	if !exists || len(location) != 1 {
		t.Fatalf("invalid location header: %v", location)
	}
	got, err := s.LockAndLoad(strings.TrimLeft(location[0], "/"))
	if err != nil {
		t.Errorf("loading from store failed: %v", err)
	}
	want := domain.NewSession()
	want.Choices = []string{"one", "two"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrong session in store. got %v want %v", got, want)
	}
}

func TestChoices(t *testing.T) {
	s := store.New()
	s.Create("id", sessionWithChoices("alice", "bob", "carol"))

	req, err := http.NewRequest("GET", "/id", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	router := handler.NewRouter(s, nil)
	router.ServeHTTP(rr, req)

	// returns ok
	if got, want := rr.Code, http.StatusOK; got != want {
		t.Errorf("wrong return code. got %v want %v", got, want)
	}
	want := handler.ChoicesResponse{
		Choices: []string{"alice", "bob", "carol"},
		Open: false,
	}
	var got handler.ChoicesResponse
	json.Unmarshal(rr.Body.Bytes(), &got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrong response. got %v want %v", got, want)
	}
}

func sessionWithChoices(choices ...string) *domain.Session {
	res := domain.NewSession()
	res.Choices = choices
	return res
}
