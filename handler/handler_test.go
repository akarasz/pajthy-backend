package handler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/event"
	"github.com/akarasz/pajthy-backend/handler"
	"github.com/akarasz/pajthy-backend/store"
)

func TestCreateSession(t *testing.T) {
	s := store.New()
	r := handler.New(s, nil)

	rr := newRequest(t, r, "POST", "/", `["one", "two"]`)

	// returns created
	assert.Exactly(t, http.StatusCreated, rr.Code)

	// id in location header where the session is saved in the store
	location, exists := rr.HeaderMap["Location"]
	if !exists || len(location) != 1 {
		t.Fatalf("invalid location header: %v", location)
	}

	if assert.Contains(t, rr.HeaderMap, "Location") && assert.Len(t, rr.HeaderMap["Location"], 1) {
		got := readFromStore(t, s, strings.TrimLeft(rr.HeaderMap["Location"][0], "/"))
		assert.Exactly(t, sessionWithChoices("one", "two"), got)
	}
}

func TestChoices(t *testing.T) {
	s := store.New()
	r := handler.New(s, nil)

	// requesting a nonexistent one return 404
	r1 := newRequest(t, r, "GET", "/id", nil)
	assert.Exactly(t, http.StatusNotFound, r1.Code)

	// successful request
	insertToStore(t, s, "id", sessionWithChoices("alice", "bob", "carol"))

	r2 := newRequest(t, r, "GET", "/id", nil)
	assert.Exactly(t, http.StatusOK, r2.Code)
	assert.JSONEq(t, `{
			"Choices": ["alice", "bob", "carol"],
			"Open": false
		}`,
		r2.Body.String())
}

func TestVote(t *testing.T) {
	s := store.New()
	e := event.New()
	r := handler.New(s, e)

	closed := sessionWithChoices("red", "blue")
	insertToStore(t, s, "closed", closed)

	open := sessionWithChoices("red", "blue")
	open.Participants = []string{"Alice", "Bob"}
	open.Open = true
	insertToStore(t, s, "open", open)

	// requesting nonexisting session should return 404
	r1 := newRequest(t, r, "PUT", "/notexists", `{"Choice": "red", "Participant": "Alice"}`)
	assert.Exactly(t, http.StatusNotFound, r1.Code)

	// voting in a closed session returns 400
	r2 := newRequest(t, r, "PUT", "/closed", `{"Choice": "red", "Participant": "Alice"}`)
	assert.Exactly(t, http.StatusBadRequest, r2.Code)
	assert.Exactly(t, "session is closed\n", r2.Body.String())

	// voting as a nonparticipant return 400
	r3 := newRequest(t, r, "PUT", "/open", `{"Choice": "red", "Participant": "Carol"}`)
	assert.Exactly(t, http.StatusBadRequest, r3.Code)
	assert.Exactly(t, "not a valid participant\n", r3.Body.String())

	// voting to a nonexisting option returns 400
	r4 := newRequest(t, r, "PUT", "/open", `{"Choice": "green", "Participant": "Alice"}`)
	assert.Exactly(t, http.StatusBadRequest, r4.Code)
	assert.Exactly(t, "not a valid choice\n", r4.Body.String())

	cEvents, vEvents := subscribe(t, e, "open", 3, 1)

	// successful vote, waiting for more
	r5 := newRequest(t, r, "PUT", "/open", `{"Choice": "red", "Participant": "Alice"}`)
	assert.Exactly(t, http.StatusAccepted, r5.Code)

	if current := readFromStore(t, s, "open"); assert.Contains(t, current.Votes, "Alice") {
		assert.Exactly(t, "red", current.Votes["Alice"])
		assert.True(t, current.Open)
	}

	if got := <-cEvents; assert.NotNil(t, got) {
		assert.Exactly(t, event.Vote, got.Kind)
		assert.Exactly(t,
			&handler.VotesChangedData{
				Votes: map[string]string{"Alice": "red"},
			},
			got.Data)
	}

	// successful vote, last one
	r6 := newRequest(t, r, "PUT", "/open", `{"Choice": "blue", "Participant": "Bob"}`)
	assert.Exactly(t, http.StatusAccepted, r6.Code)

	if current := readFromStore(t, s, "open"); assert.Contains(t, current.Votes, "Bob") {
		assert.Exactly(t, "blue", current.Votes["Bob"])
		assert.False(t, current.Open)
	}

	if got := <-vEvents; assert.NotNil(t, got) {
		assert.Exactly(t, event.Disabled, got.Kind)
	}
	if got := <-cEvents; assert.NotNil(t, got) {
		assert.Exactly(t, event.Disabled, got.Kind)
	}
	if got := <-cEvents; assert.NotNil(t, got) {
		assert.Exactly(t, event.Vote, got.Kind)
		assert.Exactly(t,
			&handler.VotesChangedData{
				Votes: map[string]string{
					"Alice": "red",
					"Bob":   "blue",
				},
			},
			got.Data)
	}
}

func TestGetSession(t *testing.T) {
	s := store.New()
	r := handler.New(s, nil)

	// returns 404 when no id is in store
	r1 := newRequest(t, r, "GET", "/abcde/control", nil)
	assert.Exactly(t, http.StatusNotFound, r1.Code)

	// successful request
	insertToStore(t, s, "abcde", sessionWithChoices("yes", "no"))

	r2 := newRequest(t, r, "GET", "/abcde/control", nil)
	assert.Exactly(t, http.StatusOK, r2.Code)
	assert.JSONEq(t, `{
			"Choices": ["yes", "no"],
			"Participants": [],
			"Votes": {},
			"Open": false
		}`, r2.Body.String())
}

func TestStartVote(t *testing.T) {
	s := store.New()
	e := event.New()
	r := handler.New(s, e)

	// returns 404 when no id is in store
	r1 := newRequest(t, r, "PATCH", "/bcdef/control/start", nil)
	assert.Exactly(t, http.StatusNotFound, r1.Code)

	// successful request
	insertToStore(t, s, "bcdef", &domain.Session{
		Choices:      []string{"dog", "cat"},
		Open:         false,
		Votes:        map[string]string{"Alice": "dog"},
		Participants: []string{"Alice"},
	})

	controllerEvent, voterEvent := subscribe(t, e, "bcdef", 1, 1)

	r2 := newRequest(t, r, "PATCH", "/bcdef/control/start", nil)
	assert.Exactly(t, http.StatusAccepted, r2.Code)

	sess := readFromStore(t, s, "bcdef")
	assert.True(t, sess.Open)
	assert.Empty(t, sess.Votes)
	if got := <-voterEvent; assert.NotNil(t, got) {
		assert.Exactly(t, event.Enabled, got.Kind)
	}
	if got := <-controllerEvent; assert.NotNil(t, got) {
		assert.Exactly(t, event.Enabled, got.Kind)
	}
}

func TestStopVote(t *testing.T) {
	s := store.New()
	e := event.New()
	r := handler.New(s, e)

	// returns 404 when no id is in store
	r1 := newRequest(t, r, "PATCH", "/bcdef/control/stop", nil)
	assert.Exactly(t, http.StatusNotFound, r1.Code)

	// successful request
	insertToStore(t, s, "bcdef", &domain.Session{
		Choices:      []string{"dog", "cat"},
		Open:         true,
		Votes:        map[string]string{"Alice": "dog"},
		Participants: []string{"Alice"},
	})

	controllerEvent, voterEvent := subscribe(t, e, "bcdef", 1, 1)

	r2 := newRequest(t, r, "PATCH", "/bcdef/control/stop", nil)
	assert.Exactly(t, http.StatusAccepted, r2.Code)

	sess := readFromStore(t, s, "bcdef")
	assert.False(t, sess.Open)
	if got := <-voterEvent; assert.NotNil(t, got) {
		assert.Exactly(t, event.Disabled, got.Kind)
	}
	if got := <-controllerEvent; assert.NotNil(t, got) {
		assert.Exactly(t, event.Disabled, got.Kind)
	}
}

func TestResetVote(t *testing.T) {
	s := store.New()
	e := event.New()
	r := handler.New(s, e)

	// returns 404 when no id is in store
	r1 := newRequest(t, r, "PATCH", "/bcdef/control/reset", nil)
	assert.Exactly(t, http.StatusNotFound, r1.Code)

	// successful request
	insertToStore(t, s, "bcdef", &domain.Session{
		Choices:      []string{"dog", "cat"},
		Open:         true,
		Votes:        map[string]string{"Alice": "dog"},
		Participants: []string{"Alice"},
	})
	controllerEvent, voterEvent := subscribe(t, e, "bcdef", 1, 2)

	r2 := newRequest(t, r, "PATCH", "/bcdef/control/reset", nil)
	assert.Exactly(t, http.StatusAccepted, r2.Code)

	sess := readFromStore(t, s, "bcdef")
	assert.False(t, sess.Open)
	assert.Empty(t, sess.Votes)
	if got := <-voterEvent; assert.NotNil(t, got) {
		assert.Exactly(t, event.Reset, got.Kind)
	}
	if got := <-controllerEvent; assert.NotNil(t, got) {
		assert.Exactly(t, event.Reset, got.Kind)
	}
	if got := <-controllerEvent; assert.NotNil(t, got) {
		assert.Exactly(t, event.Vote, got.Kind)
		assert.Empty(t, got.Data.(*handler.VotesChangedData).Votes)
	}
}

func TestKickParticipant(t *testing.T) {
	s := store.New()
	e := event.New()
	r := handler.New(s, e)

	// returns 404 when no id is in store
	r1 := newRequest(t, r, "PATCH", "/bcdef/control/kick", `"Bob"`)
	assert.Exactly(t, http.StatusNotFound, r1.Code)

	// successful request
	insertToStore(t, s, "bcdef", &domain.Session{
		Choices:      []string{"square", "circle", "triangle"},
		Open:         false,
		Votes:        map[string]string{},
		Participants: []string{"Alice", "Bob"},
	})
	controllerEvent, _ := subscribe(t, e, "bcdef", 1, 0)

	r2 := newRequest(t, r, "PATCH", "/bcdef/control/kick", `"Bob"`)
	assert.Exactly(t, http.StatusNoContent, r2.Code)

	sess := readFromStore(t, s, "bcdef")
	assert.NotContains(t, sess.Participants, "Bob")
	if got := <-controllerEvent; assert.NotNil(t, got) {
		assert.Exactly(t, event.ParticipantsChange, got.Kind)
		assert.Exactly(t, sess.Participants, got.Data.(*handler.ParticipantsChangedData).Participants)
	}
}

func TestControlWS(t *testing.T) {
	s := store.New()
	e := event.New()
	server := httptest.NewServer(handler.New(s, e))
	defer server.Close()
	baseUrl := "ws" + strings.TrimPrefix(server.URL, "http")

	insertToStore(t, s, "aaaaa", sessionWithChoices("morning", "evening"))

	ws, _, err := websocket.DefaultDialer.Dial(baseUrl+"/aaaaa/control/ws", nil)
	require.NoError(t, err)
	defer ws.Close()

	e.Emit("aaaaa", event.Controller, event.Enabled, nil)
	e.Emit("aaaaa", event.Voter, event.Disabled, nil)

	_, p, err := ws.ReadMessage()
	assert.NoError(t, err)
	assert.JSONEq(t, `{"Kind": "enabled", "Data": null}`, string(p))
}

func TestJoin(t *testing.T) {
	s := store.New()
	e := event.New()
	r := handler.New(s, e)

	// requesting nonexisting session should return 404
	r1 := newRequest(t, r, "PUT", "/ididi/join", `"Alice"`)
	assert.Exactly(t, http.StatusNotFound, r1.Code)

	// successful request
	insertToStore(t, s, "ididi", &domain.Session{
		Choices:      []string{"dog", "cat"},
		Open:         false,
		Votes:        map[string]string{},
		Participants: []string{},
	})
	controllerEvent, _ := subscribe(t, e, "ididi", 1, 0)

	r2 := newRequest(t, r, "PUT", "/ididi/join", `"Alice"`)
	assert.Exactly(t, http.StatusCreated, r2.Code)

	sess := readFromStore(t, s, "ididi")
	assert.Contains(t, sess.Participants, "Alice")

	if got := <-controllerEvent; assert.NotNil(t, got) {
		assert.Exactly(t, event.ParticipantsChange, got.Kind)
		assert.Exactly(t, sess.Participants, got.Data.(*handler.ParticipantsChangedData).Participants)
	}

	// joining with an existing name returns 409
	r3 := newRequest(t, r, "PUT", "/ididi/join", `"Alice"`)
	assert.Exactly(t, http.StatusConflict, r3.Code)
}

func TestWS(t *testing.T) {
	s := store.New()
	e := event.New()
	server := httptest.NewServer(handler.New(s, e))
	defer server.Close()
	baseUrl := "ws" + strings.TrimPrefix(server.URL, "http")

	insertToStore(t, s, "aaaaa", sessionWithChoices("morning", "evening"))

	ws, _, err := websocket.DefaultDialer.Dial(baseUrl+"/aaaaa/ws", nil)
	require.NoError(t, err)
	defer ws.Close()

	e.Emit("aaaaa", event.Controller, event.Enabled, nil)
	e.Emit("aaaaa", event.Voter, event.Disabled, nil)

	_, p, err := ws.ReadMessage()
	assert.NoError(t, err)
	assert.JSONEq(t, `{"Kind": "disabled", "Data": null}`, string(p))
}

func sessionWithChoices(choices ...string) *domain.Session {
	res := domain.NewSession()
	res.Choices = choices
	return res
}

func insertToStore(t *testing.T, s *store.Store, id string, session *domain.Session) {
	require.NoError(t, s.Create(id, session))
}

func readFromStore(t *testing.T, s *store.Store, id string) *domain.Session {
	res, err := s.Load(id)
	assert.NoError(t, err)
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
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	return rr
}

func waitForEvent(t *testing.T, e *event.Event, done chan bool, id string, r event.Role, result chan *event.Payload) {
	c, err := e.Subscribe(id, r, id)
	assert.NoError(t, err)

	done <- true

	for {
		select {
		case got := <-c:
			result <- got
		case <-time.After(1 * time.Second):
			result <- nil
			close(result)
			return
		}
	}
}

func subscribe(t *testing.T, e *event.Event, id string, expectedControllerEvents, expectedVoterEvents int) (chan *event.Payload, chan *event.Payload) {
	controllerEvent := make(chan *event.Payload, expectedControllerEvents)
	voterEvent := make(chan *event.Payload, expectedVoterEvents)

	done := make(chan bool, 2)

	if expectedVoterEvents > 0 {
		go waitForEvent(t, e, done, id, event.Voter, voterEvent)
	} else {
		done <- true
	}

	if expectedControllerEvents > 0 {
		go waitForEvent(t, e, done, id, event.Controller, controllerEvent)
	} else {
		done <- true
	}

	<-done
	<-done

	return controllerEvent, voterEvent
}
