package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/domain"
)

func CreateSession(w http.ResponseWriter, r *http.Request) {
	log.Print("create session")

	w.Header().Set("Location", fmt.Sprintf("/abcde"))
	w.WriteHeader(http.StatusCreated)
}

func Choices(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("choices %q", session)

	res := []string{"option #1", "option #2", "option #3"}

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

	w.WriteHeader(http.StatusAccepted)
}

func GetSession(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("get session %q", session)

	res := domain.Session{
		Choices:      []string{},
		Participants: []string{},
		Votes:        []domain.Vote{},
		Open:         false,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "response json encoding", http.StatusInternalServerError)
		return
	}
}

func StartVote(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("start vote %q", session)

	w.WriteHeader(http.StatusAccepted)
}

func ResetVote(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("reset vote %q", session)

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

	w.WriteHeader(http.StatusNoContent)
}

func ControlWS(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("control ws %q", session)

	w.WriteHeader(http.StatusOK)
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

	w.WriteHeader(http.StatusCreated)
}

func WS(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("ws %q", session)

	w.WriteHeader(http.StatusOK)
}
