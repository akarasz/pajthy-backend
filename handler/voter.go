package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/domain"
)

type choicesResponse struct {
	Choices []string
	Open    bool
}

func (h *Handler) choices(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	log.Printf("choices %q", session)

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	res := &choicesResponse{
		Choices: s.Choices,
		Open:    s.Open,
	}

	if err := sendJSON(w, res); err != nil {
		return
	}
}

func (h *Handler) vote(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	var v domain.Vote
	if err := parseContent(w, r, &v); err != nil {
		return
	}

	log.Printf("vote %q %q", session, v)

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	if !s.Open {
		http.Error(w, "session is closed", http.StatusBadRequest)
		return
	}

	hasVoter := false
	for _, p := range s.Participants {
		if p == v.Participant {
			hasVoter = true
			break
		}
	}
	if !hasVoter {
		http.Error(w, "invalid voter", http.StatusBadRequest)
		return
	}

	hasChoice := false
	for _, c := range s.Choices {
		if c == v.Choice {
			hasChoice = true
			break
		}
	}
	if !hasChoice {
		http.Error(w, "invalid choice", http.StatusBadRequest)
		return
	}

	s.Votes[v.Participant] = v.Choice

	if len(s.Votes) == len(s.Participants) {
		s.Open = false
		h.emitVoteDisabled(session)
	}

	if err = h.store.Update(session, s); err != nil {
		handleStoreError(w, err)
		return
	}

	h.emitVote(session, s.Votes)

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) join(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	var name string
	if err := parseContent(w, r, &name); err != nil {
		return
	}

	log.Printf("join %q %q", session, name)

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	for _, p := range s.Participants {
		if p == name {
			http.Error(w, "already joined", http.StatusConflict)
			return
		}
	}
	s.Participants = append(s.Participants, name)

	err = h.store.Update(session, s)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	h.emitParticipantsChange(session, s.Participants)

	w.WriteHeader(http.StatusCreated)
}
