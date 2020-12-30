package event_test

import (
	"testing"
	"time"

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
	if got := receiveSinglePayload(c); got == nil {
		t.Errorf("no event received")
	}
}

func TestUnsubscribe(t *testing.T) {
	e := event.New()

	// after unsubscribe emitted events are not showing on channel
	c, err := e.Subscribe("id", event.Voter, "wsID")
	if err != nil {
		t.Fatalf("unexpected error during subscribe: %v", err)
	}
	if err := e.Unsubscribe("id", "wsID"); err != nil {
		t.Fatalf("unexpected error during unsubscribe: %v", err)
	}
	go e.Emit("id", event.Voter, event.Enabled, "payload")
	if got := receiveSinglePayload(c); got != nil {
		t.Errorf("event received after unsubscribe")
	}

	// unsubscribing with uknown key returns an error
	if err := e.Unsubscribe("id", "wsID"); err == nil {
		t.Fatalf("no error returned when unsubscribing a non-existent ws")
	}
}

func TestEmit(t *testing.T) {
	// emitting an event will send a payload to all session subscription for role

	// type shows up in the payload

	// body shows up in the payload
}

func receiveSinglePayload(c chan *event.Payload) *event.Payload {
	done := make(chan *event.Payload)
	go func() {
		select {
		case got := <-c:
			done <- got
		case <-time.After(1 * time.Second):
			done <- nil
		}
	}()
	return <-done
}
