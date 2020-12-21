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

func CreateSession(w http.ResponseWriter, r *http.Request) {
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

	id, err := controller.CreateSession(body)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/%s", id))
	w.WriteHeader(http.StatusCreated)
}

func Choices(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("choices %q", session)

	res, err := controller.Choices(session)
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

func Vote(w http.ResponseWriter, r *http.Request) {
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

	err = controller.Vote(session, &body)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func GetSession(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("get session %q", session)

	res, err := controller.GetSession(session)
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

type SessionRes struct {
	Choices      []string
	Participants []string
	Votes        []*domain.Vote
	Open         bool
}

func StartVote(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("start vote %q", session)

	err := controller.StartVote(session)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func ResetVote(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("reset vote %q", session)

	err := controller.ResetVote(session)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func KickParticipant(w http.ResponseWriter, r *http.Request) {
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

	err = controller.KickParticipant(session, body)
	if err != nil {
		handleControllerError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func Join(w http.ResponseWriter, r *http.Request) {
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

	err = controller.Join(session, body)
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
