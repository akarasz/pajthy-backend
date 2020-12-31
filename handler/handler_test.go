package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/handler"
	"github.com/akarasz/pajthy-backend/store"
	"github.com/gorilla/mux"
)

func TestCreateSession(t *testing.T) {
	s := store.New()
	r := handler.NewRouter(s, nil)

	rr := newRequest(t, r, "POST", "/", `["one", "two"]`)

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
	r := handler.NewRouter(s, nil)

	// requesting a nonexistent one return 404
	rr := newRequest(t, r, "GET", "/not-existing", nil)

	if got, want := rr.Code, http.StatusNotFound; got != want {
		t.Errorf("wrong return code for not-existing id. got %v want %v", got, want)
	}

	// successful request
	s.Create("id", sessionWithChoices("alice", "bob", "carol"))

	rr = newRequest(t, r, "GET", "/id", nil)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Errorf("wrong return code for existing id. got %v want %v", got, want)
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

func TestVote(t *testing.T) {
	// requesting with invalid body should return 400
	// requesting nonexisting session should return 404
	// voting in a closed session returns 400
	// voting as a nonparticipant return 400
	// voting to a nonexisting option returns 400
	// successful vote returns 201
	// successful vote saves the vote
	// successful vote emits a vote event to controllers
	// successful vote closes the vote if all votes are in
	// successful vote emits a disabled event if all votes are in to controllers and votes
}

func TestGetSession(t *testing.T) {
	// returns 404 when no id is in store
	// successful request returns 200
	// successful request returns session object
}

func TestStartVote(t *testing.T) {
	// returns 404 when no id is in store
	// returns 202 when successful
	// sets state to open in session
	// clears votes in session
	// emits vote enabled to controllers and voters
}

func TestStopVote(t *testing.T) {
	// returns 404 when no id is in store
	// returns 202 when successful
	// sets state to closed in session
	// votes in session remaining
	// emits vote disabled to controllers and voters
}

func TestResetVote(t *testing.T) {
	// returns 404 when no id is in store
	// returns 202 when successful
	// sets state to closed in session
	// clears votes in session
	// emits vote disabled to controllers and voters
}

func TestKickParticipant(t *testing.T) {
	// returns 404 when no id is in store
	// returns 400 when trying to remove a nonexistsing participant
	// returns 204 when successful
	// removes the participant from the store
	// emits a participant change event to controllers
}

func TestJoin(t *testing.T) {
	// requesting with invalid body should return 400
	// requesting nonexisting session should return 404
	// joining with an existing name returns 409
	// successful vote returns 201
	// successful join adds the participant to the session
	// successful join emits a join event
}

func sessionWithChoices(choices ...string) *domain.Session {
	res := domain.NewSession()
	res.Choices = choices
	return res
}

func newRequest(t *testing.T, r *mux.Router, method string, url string, body interface{}) *httptest.ResponseRecorder {
	var reqBody io.Reader
	switch body.(type) {
	case string:
		reqBody = strings.NewReader(body.(string))
	case io.Reader:
		reqBody = body.(io.Reader)
	default:
		reqBody = nil
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	return rr
}