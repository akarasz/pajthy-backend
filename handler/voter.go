package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/domain"
)

type ChoicesResponse struct {
	Choices []string
	Open    bool
}

func (h *Handler) choices(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	log.Printf("choices %q", session)

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		showStoreError(w, err)
		return
	}

	res := &ChoicesResponse{
		Choices: s.Choices,
		Open:    s.Open,
	}

	if err := showJSON(w, res); err != nil {
		return
	}
}

func (h *Handler) vote(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	var v domain.Vote
	if err := readContent(w, r, &v); err != nil {
		return
	}

	log.Printf("vote %q %q", session, v)

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		showStoreError(w, err)
		return
	}

	if !s.Open {
		showError(w, http.StatusBadRequest, "session is closed", nil)
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
		showError(w, http.StatusBadRequest, "not a valid participant", nil)
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
		showError(w, http.StatusBadRequest, "not a valid choice", nil)
		return
	}

	s.Votes[v.Participant] = v.Choice

	if len(s.Votes) == len(s.Participants) {
		s.Open = false
		h.emitVoteDisabled(session)
	}

	if err = h.store.Update(session, s); err != nil {
		showStoreError(w, err)
		return
	}

	h.emitVote(session, s.Votes)

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) join(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	var name string
	if err := readContent(w, r, &name); err != nil {
		return
	}

	log.Printf("join %q %q", session, name)

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		showStoreError(w, err)
		return
	}

	for _, p := range s.Participants {
		if p == name {
			showError(w, http.StatusConflict, "already joined", nil)
			return
		}
	}
	s.Participants = append(s.Participants, name)

	err = h.store.Update(session, s)
	if err != nil {
		showStoreError(w, err)
		return
	}

	h.emitParticipantsChange(session, s.Participants)

	w.WriteHeader(http.StatusCreated)
}
