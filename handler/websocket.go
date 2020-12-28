package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/akarasz/pajthy-backend/event"
	"github.com/akarasz/pajthy-backend/store"
)

const (
	pongWait   = 30 * time.Second
	pingPeriod = (pongWait * 8) / 10
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) writer(ws *websocket.Conn, sessionID string, msgs <-chan *event.Payload) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		h.event.Unsubscribe(sessionID, ws)
		pingTicker.Stop()
		ws.Close()
	}()

	for {
		select {
		case msg := <-msgs:
			if err := ws.WriteJSON(msg); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (h *Handler) reader(ws *websocket.Conn, sessionID string) {
	defer func() {
		h.event.Unsubscribe(sessionID, ws)
		ws.Close()
	}()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *Handler) ws(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			http.Error(w, "unknown error", http.StatusInternalServerError)
		}
		return
	}

	if _, err := h.store.Load(session); err == store.ErrNotExists {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	c, err := h.event.Subscribe(session, event.Voter, ws)
	if err != nil {
		http.Error(w, "unable to subscribe", http.StatusInternalServerError)
		return
	}

	log.Printf("ws %q", session)

	go h.writer(ws, session, c)
	h.reader(ws, session)
}

func (h *Handler) controlWS(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			http.Error(w, "unknown error", http.StatusInternalServerError)
		}
		return
	}

	if _, err := h.store.Load(session); err == store.ErrNotExists {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	c, err := h.event.Subscribe(session, event.Controller, ws)
	if err != nil {
		http.Error(w, "unable to subscribe", http.StatusInternalServerError)
		return
	}

	log.Printf("control ws %q", session)

	go h.writer(ws, session, c)
	h.reader(ws, session)
}

type openChangedData struct {
	Open bool
}

type votesChangedData struct {
	Votes map[string]string
}

type participantsChangedData struct {
	Participants []string
}

func (h *Handler) emitVoteEnabled(id string) {
	m := &openChangedData{Open: true}
	h.event.Emit(id, event.Voter, event.Enabled, m)
	h.event.Emit(id, event.Controller, event.Enabled, m)
}

func (h *Handler) emitVoteDisabled(id string) {
	m := &openChangedData{Open: false}
	h.event.Emit(id, event.Voter, event.Disabled, m)
	h.event.Emit(id, event.Controller, event.Disabled, m)
}

func (h *Handler) emitReset(id string) {
	m := &openChangedData{Open: false}
	h.event.Emit(id, event.Voter, event.Reset, m)
	h.event.Emit(id, event.Controller, event.Reset, m)
}

func (h *Handler) emitVote(id string, votes map[string]string) {
	h.event.Emit(id, event.Controller, event.Vote, &votesChangedData{Votes: votes})
}

func (c *Handler) emitParticipantsChange(id string, participants []string) {
	c.event.Emit(
		id,
		event.Controller,
		event.ParticipantsChange,
		&participantsChangedData{Participants: participants})
}
