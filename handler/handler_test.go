package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/event"
	"github.com/akarasz/pajthy-backend/handler"
	"github.com/akarasz/pajthy-backend/store"
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
	want := sessionWithChoices("one", "two")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrong session in store. got %v want %v", got, want)
	}
}

func TestChoices(t *testing.T) {
	s := store.New()
	r := handler.NewRouter(s, nil)

	// requesting a nonexistent one return 404
	rr := newRequest(t, r, "GET", "/id", nil)

	if got, want := rr.Code, http.StatusNotFound; got != want {
		t.Errorf("wrong return code for not-existing id. got %v want %v", got, want)
	}

	// successful request
	insertToStore(t, s, "id", sessionWithChoices("alice", "bob", "carol"))

	rr = newRequest(t, r, "GET", "/id", nil)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Errorf("wrong response code. got %v want %v", got, want)
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
	s := store.New()
	r := handler.NewRouter(s, nil)

	// returns 404 when no id is in store
	rr := newRequest(t, r, "GET", "/abcde/control", nil)
	if got, want := rr.Code, http.StatusNotFound; got != want {
		t.Errorf("wrong response code for not-existing id. got %v want %v", got, want)
	}

	// successful request
	insertToStore(t, s, "abcde", sessionWithChoices("yes", "no"))

	rr = newRequest(t, r, "GET", "/abcde/control", nil)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Errorf("wrong response code. got %v want %v", got, want)
	}
	want := domain.Session{
		Choices: []string{"yes", "no"},
		Participants: []string{},
		Votes: map[string]string{},
		Open: false,
	}
	var got domain.Session
	json.Unmarshal(rr.Body.Bytes(), &got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrong response. got %v want %v", got, want)
	}
}

func TestStartVote(t *testing.T) {
	s := store.New()
	e := event.New()
	r := handler.NewRouter(s, e)

	// returns 404 when no id is in store
	rr := newRequest(t, r, "PATCH", "/bcdef/control/start", nil)
	if got, want := rr.Code, http.StatusNotFound; got != want {
		t.Errorf("wrong response code for not-existing id. got %v want %v", got, want)
	}

	// successful request
	insertToStore(t, s, "bcdef", &domain.Session{
		Choices: []string{ "dog", "cat" },
		Open: false,
		Votes: map[string]string{ "Alice": "dog" },
		Participants: []string{ "Alice" },
	})
	voterEvent := make(chan *event.Payload)
	controllerEvent := make(chan *event.Payload)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go waitForEvent(t, e, wg, "bcdef", event.Voter, voterEvent)
	go waitForEvent(t, e, wg, "bcdef", event.Controller, controllerEvent)
	wg.Wait()

	rr = newRequest(t, r, "PATCH", "/bcdef/control/start", nil)
	if got, want := rr.Code, http.StatusAccepted; got != want {
		t.Errorf("wrong response code. got %v want %v", got, want)
	}
	sess := readFromStore(t, s, "bcdef")
	if got, want := sess.Open, true; got != want {
		t.Errorf("wrong Open. got %v want %v", got, want)
	}
	if got, want := len(sess.Votes), 0; got != want {
		t.Errorf("wrong length of Votes. got %v want %v", got, want)
	}
	if got := <-voterEvent; got == nil || got.Kind != event.Enabled {
		t.Errorf("wrong payload for voters: %v", got)
	}
	if got := <-controllerEvent; got == nil || got.Kind != event.Enabled {
		t.Errorf("wrong payload for controllers: %v", got)
	}
}

func TestStopVote(t *testing.T) {
	s := store.New()
	e := event.New()
	r := handler.NewRouter(s, e)

	// returns 404 when no id is in store
	rr := newRequest(t, r, "PATCH", "/bcdef/control/stop", nil)
	if got, want := rr.Code, http.StatusNotFound; got != want {
		t.Errorf("wrong response code for not-existing id. got %v want %v", got, want)
	}

	// successful request
	insertToStore(t, s, "bcdef", &domain.Session{
		Choices: []string{ "dog", "cat" },
		Open: true,
		Votes: map[string]string{ "Alice": "dog" },
		Participants: []string{ "Alice" },
	})
	voterEvent := make(chan *event.Payload)
	controllerEvent := make(chan *event.Payload)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go waitForEvent(t, e, wg, "bcdef", event.Voter, voterEvent)
	go waitForEvent(t, e, wg, "bcdef", event.Controller, controllerEvent)
	wg.Wait()

	rr = newRequest(t, r, "PATCH", "/bcdef/control/stop", nil)
	if got, want := rr.Code, http.StatusAccepted; got != want {
		t.Errorf("wrong response code. got %v want %v", got, want)
	}
	sess := readFromStore(t, s, "bcdef")
	if got, want := sess.Open, false; got != want {
		t.Errorf("wrong Open. got %v want %v", got, want)
	}
	if got := <-voterEvent; got == nil || got.Kind != event.Disabled {
		t.Errorf("wrong payload for voters: %v", got)
	}
	if got := <-controllerEvent; got == nil || got.Kind != event.Disabled {
		t.Errorf("wrong payload for controllers: %v", got)
	}
}

func TestResetVote(t *testing.T) {
	s := store.New()
	e := event.New()
	r := handler.NewRouter(s, e)

	// returns 404 when no id is in store
	rr := newRequest(t, r, "PATCH", "/bcdef/control/reset", nil)
	if got, want := rr.Code, http.StatusNotFound; got != want {
		t.Errorf("wrong response code for not-existing id. got %v want %v", got, want)
	}

	// successful request
	insertToStore(t, s, "bcdef", &domain.Session{
		Choices: []string{ "dog", "cat" },
		Open: true,
		Votes: map[string]string{ "Alice": "dog" },
		Participants: []string{ "Alice" },
	})
	voterEvent := make(chan *event.Payload)
	controllerEvent := make(chan *event.Payload, 2)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go waitForEvent(t, e, wg, "bcdef", event.Voter, voterEvent)
	go waitForEvent(t, e, wg, "bcdef", event.Controller, controllerEvent)
	wg.Wait()

	rr = newRequest(t, r, "PATCH", "/bcdef/control/reset", nil)
	if got, want := rr.Code, http.StatusAccepted; got != want {
		t.Errorf("wrong response code. got %v want %v", got, want)
	}
	sess := readFromStore(t, s, "bcdef")
	if got, want := sess.Open, false; got != want {
		t.Errorf("wrong Open. got %v want %v", got, want)
	}
	if got, want := len(sess.Votes), 0; got != want {
		t.Errorf("wrong length of Votes. got %v want %v", got, want)
	}
	if got := <-voterEvent; got == nil || got.Kind != event.Reset {
		t.Errorf("wrong payload for voters when reset: %v", got)
	}
	if got := <-controllerEvent; got == nil || got.Kind != event.Reset {
		t.Errorf("wrong payload for controllers when reset: %v", got)
	}
	if got := <-controllerEvent; got == nil || got.Kind != event.Vote ||
			len(got.Data.(*handler.VotesChangedData).Votes) > 0 {
		t.Errorf("wrong payload for controllers when Vote: %v", got)
	}
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

func insertToStore(t *testing.T, s *store.Store, id string, session *domain.Session) {
	if err := s.Create(id, session); err != nil {
		t.Fatal(err)
	}
}

func readFromStore(t *testing.T, s *store.Store, id string) *domain.Session {
	res, err := s.LockAndLoad(id)
	defer s.Unlock(id)
	if err != nil {
		t.Fatal(err)
	}
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

func waitForEvent(t *testing.T, e *event.Event, wg *sync.WaitGroup, id string, r event.Role, result chan *event.Payload) {
	c, err := e.Subscribe(id, r, id)
	if err != nil {
		t.Fatal(err)
	}
	wg.Done()
	for {
		select {
		case got := <-c:
			result <- got
		case <- time.After(1 * time.Second):
			result <- nil
			return
		}
	}
}