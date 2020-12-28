package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/event"
	"github.com/akarasz/pajthy-backend/store"
)

type Handler struct {
	store store.Store
	event event.Event
}

func NewRouter(s store.Store, e event.Event) *mux.Router {
	h := &Handler{
		store: s,
		event: e,
	}
	r := mux.NewRouter()

	r.Use(corsMiddleware)
	r.HandleFunc("/", h.createSession).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{session}", h.choices).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}", h.vote).
		Methods("PUT", "OPTIONS")
	r.HandleFunc("/{session}/control", h.getSession).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}/control/start", h.startVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/stop", h.stopVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/reset", h.resetVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/kick", h.kickParticipant).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/ws", h.controlWS).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}/join", h.join).
		Methods("PUT", "OPTIONS")
	r.HandleFunc("/{session}/ws", h.ws).
		Methods("GET", "OPTIONS")

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseContent(w http.ResponseWriter, r *http.Request, dest interface{}) error {
	rawBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "wrong body", http.StatusBadRequest)
		return err
	}
	if err := json.Unmarshal(rawBody, &dest); err != nil {
		http.Error(w, fmt.Sprintf("request json decoding: %v", err), http.StatusBadRequest)
		return err
	}
	return nil
}

func sendJSON(w http.ResponseWriter, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "response json encoding", http.StatusInternalServerError)
		return err
	}
	return nil
}

func handleStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotExists) {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
	} else {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}
}
