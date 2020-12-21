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
	r.Use(handler.CorsMiddleware)
	r.HandleFunc("/", handler.CreateSession).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{session}", handler.Choices).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}", handler.Vote).
		Methods("PUT", "OPTIONS")
	r.HandleFunc("/{session}/control", handler.GetSession).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}/control/start", handler.StartVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/reset", handler.ResetVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/kick", handler.KickParticipant).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/ws", handler.ControlWS).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}/join", handler.Join).
		Methods("PUT", "OPTIONS")
	r.HandleFunc("/{session}/ws", handler.WS).
		Methods("GET", "OPTIONS")

	log.Fatal(http.ListenAndServe(":8000", r))
}
