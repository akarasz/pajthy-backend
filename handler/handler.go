package handler

import (
	"encoding/json"
	"net/http"
	"io/ioutil"
	"fmt"
	"log"

	"github.com/gorilla/mux"
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

	res := "res"

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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}
	bodyString := string(body)

	log.Printf("vote %q %q", session, bodyString)

	w.WriteHeader(http.StatusAccepted)
}

func GetSession(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Printf("get session %q", session)

	res := "res"

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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}
	bodyString := string(body)

	log.Printf("kick participant %q %q", session, bodyString)

	w.WriteHeader(http.StatusNoContent)
}

func ControlWS(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	log.Print("control ws", session)

	w.WriteHeader(http.StatusOK)
}

func Join(w http.ResponseWriter, r *http.Request) {
	session, ok := mux.Vars(r)["session"]
	if !ok {
		http.Error(w, "wrong session", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}
	bodyString := string(body)

	log.Printf("join %q %q", session, bodyString)

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