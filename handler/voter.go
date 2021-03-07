package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
)

type ChoicesResponse struct {
	Choices []string
	Open    bool
}

func (h *Handler) choices(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	log.Printf("choices %q", session)

	ss, err := h.store.Load(session)
	if err != nil {
		showStoreError(w, err)
		return
	}
	s := ss.Data

	res := &ChoicesResponse{
		Choices: s.Choices,
		Open:    s.Open,
	}

	if err := showJSON(w, res); err != nil {
		return
	}
}

func (h *Handler) vote(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["session"]

	var v domain.Vote
	if err := readContent(w, r, &v); err != nil {
		return
	}

	log.Printf("vote %q %q", id, v)

	saved, err := store.ReadModifyWrite(id, h.store, func(s *domain.Session) (*domain.Session, error) {
		if !s.Open {
			return nil, errClosedSession
		}

		hasVoter := false
		for _, p := range s.Participants {
			if p == v.Participant {
				hasVoter = true
				break
			}
		}
		if !hasVoter {
			return nil, errInvalidParticipant
		}

		hasChoice := false
		for _, c := range s.Choices {
			if c == v.Choice {
				hasChoice = true
				break
			}
		}
		if !hasChoice {
			return nil, errInvalidChoice
		}

		s.Votes[v.Participant] = v.Choice

		s.Open = len(s.Votes) < len(s.Participants)

		return s, nil
	})

	switch err {
	case nil:
		w.WriteHeader(http.StatusAccepted)
	case errClosedSession:
		showError(w, http.StatusBadRequest, "session is closed", nil)
		return
	case errInvalidParticipant:
		showError(w, http.StatusBadRequest, "not a valid participant", nil)
		return
	case errInvalidChoice:
		showError(w, http.StatusBadRequest, "not a valid choice", nil)
		return
	case store.ErrVersionMismatch:
		showError(w, http.StatusInternalServerError, "locking error, try again later", nil)
		return
	default:
		showStoreError(w, err)
		return
	}

	h.emitVote(id, saved.Votes)
	if !saved.Open {
		h.emitVoteDisabled(id)
	}
}

func (h *Handler) join(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["session"]

	var name string
	if err := readContent(w, r, &name); err != nil {
		return
	}

	log.Printf("join %q %q", id, name)

	saved, err := store.ReadModifyWrite(id, h.store, func(s *domain.Session) (*domain.Session, error) {
		for _, p := range s.Participants {
			if p == name {
				return nil, errAlreadyJoined
			}
		}
		s.Participants = append(s.Participants, name)

		return s, nil
	})

	switch err {
	case nil:
		w.WriteHeader(http.StatusCreated)
	case errAlreadyJoined:
		showError(w, http.StatusConflict, "already joined", nil)
		return
	case store.ErrVersionMismatch:
		showError(w, http.StatusInternalServerError, "locking error, try again later", nil)
		return
	default:
		showStoreError(w, err)
		return
	}

	h.emitParticipantsChange(id, saved.Participants)
}
