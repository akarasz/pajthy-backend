package main

import (
	"log"
	"net/http"
	"math/rand"
	"time"

	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/handler"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	r := mux.NewRouter()

	r.HandleFunc("/", handler.CreateSession).
		Methods("POST")
	r.HandleFunc("/{session}", handler.Choices).
		Methods("GET")
	r.HandleFunc("/{session}", handler.Vote).
		Methods("PUT")
	r.HandleFunc("/{session}/control", handler.GetSession).
		Methods("GET")
	r.HandleFunc("/{session}/control/start", handler.StartVote).
		Methods("PATCH")
	r.HandleFunc("/{session}/control/reset", handler.ResetVote).
		Methods("PATCH")
	r.HandleFunc("/{session}/control/kick", handler.KickParticipant).
		Methods("PATCH")
	r.HandleFunc("/{session}/control/ws", handler.ControlWS).
		Methods("GET")
	r.HandleFunc("/{session}/join", handler.Join).
		Methods("PUT")
	r.HandleFunc("/{session}/ws", handler.WS).
		Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", r))
}
