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

	s, err := h.store.Load(session)
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

	err := store.OptimisticLocking(func() error {
		s, err := h.store.Load(session)
		if err != nil {
			return err
		}

		if !s.Open {
			return errClosedSession
		}

		hasVoter := false
		for _, p := range s.Participants {
			if p == v.Participant {
				hasVoter = true
				break
			}
		}
		if !hasVoter {
			return errInvalidParticipant
		}

		hasChoice := false
		for _, c := range s.Choices {
			if c == v.Choice {
				hasChoice = true
				break
			}
		}
		if !hasChoice {
			return errInvalidChoice
		}

		s.Votes[v.Participant] = v.Choice

		if len(s.Votes) == len(s.Participants) {
			s.Open = false
			h.emitVoteDisabled(session)
		}

		if err = h.store.Update(session, s); err != nil {
			return err
		}

		h.emitVote(session, s.Votes)

		return nil
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
	default:
		showStoreError(w, err)
		return
	}
}

func (h *Handler) join(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	var name string
	if err := readContent(w, r, &name); err != nil {
		return
	}

	log.Printf("join %q %q", session, name)

	err := store.OptimisticLocking(func() error {
		s, err := h.store.Load(session)
		if err != nil {
			return err
		}

		for _, p := range s.Participants {
			if p == name {
				return errAlreadyJoined
			}
		}
		s.Participants = append(s.Participants, name)

		err = h.store.Update(session, s)
		if err != nil {
			return err
		}

		h.emitParticipantsChange(session, s.Participants)

		return nil
	})

	switch err {
	case nil:
		w.WriteHeader(http.StatusCreated)
	case errAlreadyJoined:
		showError(w, http.StatusConflict, "already joined", nil)
		return
	default:
		showStoreError(w, err)
		return
	}
}
