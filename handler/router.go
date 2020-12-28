package handler

import "github.com/gorilla/mux"

func NewPajthyRouter() *mux.Router {
	r := mux.NewRouter()

	r.Use(CorsMiddleware)

	r.HandleFunc("/", CreateSession).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{session}", Choices).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}", Vote).
		Methods("PUT", "OPTIONS")
	r.HandleFunc("/{session}/control", GetSession).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}/control/start", StartVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/stop", StopVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/reset", ResetVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/kick", KickParticipant).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/ws", ControlWS).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}/join", Join).
		Methods("PUT", "OPTIONS")
	r.HandleFunc("/{session}/ws", WS).
		Methods("GET", "OPTIONS")

	return r
}
