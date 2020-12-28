package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/controller"
	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
)

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	log.Print("create session")

	var body []string
	rawBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(rawBody, &body); err != nil {
		http.Error(w, "request json decoding", http.StatusBadRequest)
		return
	}

	id, err := h.controller.CreateSession(body)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/%s", id))
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Choices(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("choices %q", session)

	res, err := h.controller.Choices(session)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "response json encoding", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Vote(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	var body domain.Vote
	rawBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(rawBody, &body); err != nil {
		http.Error(w, "request json decoding", http.StatusBadRequest)
		return
	}

	log.Printf("vote %q %q", session, body)

	err = h.controller.Vote(session, &body)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("get session %q", session)

	res, err := h.controller.GetSession(session)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "response json encoding", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) StartVote(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("start vote %q", session)

	err := h.controller.StartVote(session)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) StopVote(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("stop vote %q", session)

	err := h.controller.StopVote(session)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) ResetVote(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("reset vote %q", session)

	err := h.controller.ResetVote(session)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) KickParticipant(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	rawBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}
	body := string(rawBody)

	log.Printf("kick participant %q %q", session, body)

	err = h.controller.KickParticipant(session, body)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	rawBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}
	body := string(rawBody)

	log.Printf("join %q %q", session, body)

	err = h.controller.Join(session, body)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleControllerError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotExists) {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
	} else if errors.Is(err, controller.ErrVoterAlreadyJoined) {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusConflict)
	} else {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}
}
