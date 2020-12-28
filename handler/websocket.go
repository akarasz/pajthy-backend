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

func writer(ws *websocket.Conn, sessionID string, msgs <-chan *event.Payload) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		event.Unsubscribe(sessionID, ws)
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

func reader(ws *websocket.Conn, sessionID string) {
	defer func() {
		event.Unsubscribe(sessionID, ws)
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

func WS(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}
	if _, err := store.Load(session); err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			http.Error(w, "unknown error", http.StatusInternalServerError)
		}
		return
	}

	c, err := event.Subscribe(session, event.Voter, ws)
	if err != nil {
		http.Error(w, "unable to subscribe", http.StatusInternalServerError)
		return
	}

	log.Printf("ws %q", session)

	go writer(ws, session, c)
	reader(ws, session)
}

func ControlWS(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}
	if _, err := store.Load(session); err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			http.Error(w, "unknown error", http.StatusInternalServerError)
		}
		return
	}

	c, err := event.Subscribe(session, event.Controller, ws)
	if err != nil {
		http.Error(w, "unable to subscribe", http.StatusInternalServerError)
		return
	}

	log.Printf("control ws %q", session)

	go writer(ws, session, c)
	reader(ws, session)
}
