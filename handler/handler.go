package handler

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"log"

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

func readContent(w http.ResponseWriter, r *http.Request, dest interface{}) error {
	rawBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		showError(w, http.StatusBadRequest, "wrong body", err)
		return err
	}
	if err := json.Unmarshal(rawBody, &dest); err != nil {
		showError(w, http.StatusInternalServerError, "request json decoding", err)
		return err
	}
	return nil
}

func showJSON(w http.ResponseWriter, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		showError(w, http.StatusInternalServerError, "response json encoding", err)
		return err
	}
	return nil
}

func showStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotExists) {
		showError(w, http.StatusNotFound, "session not exists", err)
	} else if errors.Is(err, store.ErrAlreadyExists) {
		showError(w, http.StatusConflict, "session already exists", err)
	} else {
		showError(w, http.StatusInternalServerError, "unknown error", err)
	}
}

func showError(w http.ResponseWriter, code int, msg string, err error) {
	http.Error(w, msg, code)
	log.Printf("%s: %v", msg, err)
}