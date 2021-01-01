package event_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/event"
)

func TestSubscribe(t *testing.T) {
	e := event.New()

	// subscribing returns a channel where events can be emitted
	c, err := e.Subscribe("id", event.Voter, "wsID")
	assert.NoError(t, err)

	go e.Emit("id", event.Voter, event.Enabled, "payload")

	assert.NotNil(t, <-receivePayload(c), "no event received")
}

func TestUnsubscribe(t *testing.T) {
	e := event.New()

	// after unsubscribe emitted events are not showing on channel
	c := mustSubscribe(t, e, "id", event.Voter, "wsID")
	err := e.Unsubscribe("id", "wsID")
	assert.NoError(t, err)

	go e.Emit("id", event.Voter, event.Enabled, "payload")
	assert.Nil(t, <-receivePayload(c), "event received after unsubscribe")

	// unsubscribing with uknown key returns an error
	err = e.Unsubscribe("id", "wsID")
	assert.Error(t, err)
}

func TestEmit_SendingTo(t *testing.T) {
	e := event.New()

	// emitting an event will send a payload to all session subscription for role
	alice := mustSubscribe(t, e, "a", event.Voter, "alice")
	bob := mustSubscribe(t, e, "a", event.Voter, "bob")
	carol := mustSubscribe(t, e, "a", event.Controller, "carol")
	dave := mustSubscribe(t, e, "b", event.Voter, "dave")
	var cAlice, cBob, cCarol, cDave chan *event.Payload
	done := make(chan bool)
	go func() {

		cAlice = receivePayload(alice)
		cBob = receivePayload(bob)
		cCarol = receivePayload(carol)
		cDave = receivePayload(dave)

		done <- true
	}()
	e.Emit("a", event.Voter, event.Enabled, "payload")
	<-done
	assert.NotNil(t, <-cAlice)
	assert.NotNil(t, <-cBob)
	assert.Nil(t, <-cCarol)
	assert.Nil(t, <-cDave)
}

func TestEmit_Payload(t *testing.T) {
	e := event.New()

	c := mustSubscribe(t, e, "id", event.Voter, "wsID")
	want := event.NewPayload(event.Vote, &domain.Vote{
		Participant: "Alice",
		Choice:      "the right one",
	})

	go e.Emit("id", event.Voter, want.Kind, want.Data)

	if got := <-receivePayload(c); assert.NotNil(t, got) {
		assert.Exactly(t, got, want)
	}
}

func mustSubscribe(t *testing.T, e *event.Event, sessionID string, r event.Role, ws interface{}) chan *event.Payload {
	res, err := e.Subscribe(sessionID, r, ws)
	assert.NoError(t, err)
	return res
}

func receivePayload(c chan *event.Payload) chan *event.Payload {
	res := make(chan *event.Payload)
	go func() {
		select {
		case got := <-c:
			res <- got
		case <-time.After(100 * time.Millisecond):
			res <- nil
			close(res)
			return
		}
	}()
	return res
}
