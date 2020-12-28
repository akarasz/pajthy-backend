package handler

import (
	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/controller"
	"github.com/akarasz/pajthy-backend/event"
	"github.com/akarasz/pajthy-backend/store"
)

type Handler struct {
	controller *controller.Controller
	store      store.Store
	event      event.Event
}

func NewRouter(s store.Store, e event.Event) *mux.Router {
	h := &Handler{
		controller: controller.New(s, e),
		store:      s,
		event:      e,
	}
	r := mux.NewRouter()

	r.Use(CorsMiddleware)
	r.HandleFunc("/", h.CreateSession).
		Methods("POST", "OPTIONS")
	r.HandleFunc("/{session}", h.Choices).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}", h.Vote).
		Methods("PUT", "OPTIONS")
	r.HandleFunc("/{session}/control", h.GetSession).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}/control/start", h.StartVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/stop", h.StopVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/reset", h.ResetVote).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/kick", h.KickParticipant).
		Methods("PATCH", "OPTIONS")
	r.HandleFunc("/{session}/control/ws", h.ControlWS).
		Methods("GET", "OPTIONS")
	r.HandleFunc("/{session}/join", h.Join).
		Methods("PUT", "OPTIONS")
	r.HandleFunc("/{session}/ws", h.WS).
		Methods("GET", "OPTIONS")

	return r
}
