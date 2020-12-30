package event_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/event"
)

func TestSubscribe(t *testing.T) {
	e := event.New()

	// subscribing returns a channel where events can be emitted
	c, err := e.Subscribe("id", event.Voter, "wsID")
	if err != nil {
		t.Fatalf("unexpected error during subscribe: %v", err)
	}
	go e.Emit("id", event.Voter, event.Enabled, "payload")
	if got := receivePayload(c); got == nil {
		t.Errorf("no event received")
	}
}

func TestUnsubscribe(t *testing.T) {
	e := event.New()

	// after unsubscribe emitted events are not showing on channel
	c := mustSubscribe(t, e, "id", event.Voter, "wsID")
	if err := e.Unsubscribe("id", "wsID"); err != nil {
		t.Fatalf("unexpected error during unsubscribe: %v", err)
	}
	go e.Emit("id", event.Voter, event.Enabled, "payload")
	if got := receivePayload(c); got != nil {
		t.Errorf("event received after unsubscribe")
	}

	// unsubscribing with uknown key returns an error
	if err := e.Unsubscribe("id", "wsID"); err == nil {
		t.Fatalf("no error returned when unsubscribing a non-existent ws")
	}
}

func TestEmit_SendingTo(t *testing.T) {
	e := event.New()

	// emitting an event will send a payload to all session subscription for role
	alice := mustSubscribe(t, e, "a", event.Voter, "alice")
	bob := mustSubscribe(t, e, "a", event.Voter, "bob")
	carol := mustSubscribe(t, e, "a", event.Controller, "carol")
	dave := mustSubscribe(t, e, "b", event.Voter, "dave")
	var aliceReceived, bobReceived, carolReceived, daveReceived *event.Payload
	done := make(chan bool)
	go func() {
		aliceReceived = receivePayload(alice)
		bobReceived = receivePayload(bob)
		carolReceived = receivePayload(carol)
		daveReceived = receivePayload(dave)

		done <- true
	}()
	e.Emit("a", event.Voter, event.Enabled, "payload")
	<-done
	if aliceReceived == nil {
		t.Errorf("alice should have received an event")
	}
	if bobReceived == nil {
		t.Errorf("bob should have received an event")
	}
	if carolReceived != nil {
		t.Errorf("carol should have not received an event")
	}
	if daveReceived != nil {
		t.Errorf("dave should have not received an event")
	}
}

func TestEmit_Payload(t *testing.T) {
	e := event.New()

	c := mustSubscribe(t, e, "id", event.Voter, "wsID")
	wantKind := event.Vote
	wantData := &domain.Vote{
		Participant: "Alice",
		Choice:      "the right one",
	}
	go e.Emit("id", event.Voter, wantKind, wantData)
	got := receivePayload(c)
	if got == nil {
		t.Fatalf("no event received")
	}
	if want := event.NewPayload(wantKind, wantData); !reflect.DeepEqual(got, want) {
		t.Fatalf("wrong payload. got %v want %v", got, want)
	}
}

func mustSubscribe(t *testing.T, e *event.Event, sessionID string, r event.Role, ws interface{}) chan *event.Payload {
	res, err := e.Subscribe(sessionID, r, ws)
	if err != nil {
		t.Fatalf("unexpected error during subscribe: %v", err)
	}
	return res
}

func receivePayload(c chan *event.Payload) *event.Payload {
	done := make(chan *event.Payload)
	go func() {
		select {
		case got := <-c:
			done <- got
		case <-time.After(100 * time.Millisecond):
			done <- nil
		}
	}()
	return <-done
}
