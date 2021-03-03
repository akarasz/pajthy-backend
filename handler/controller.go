package handler

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
)

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	log.Print("create session")

	var choices []string
	if err := readContent(w, r, &choices); err != nil {
		return
	}

	id := generateID()
	s := domain.NewSession()
	s.Choices = choices

	if err := h.store.Save(id, s); err != nil {
		showStoreError(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/%s", id))
	w.WriteHeader(http.StatusCreated)
}

func generateID() string {
	const (
		idCharset = "abcdefghijklmnopqrstvwxyz0123456789"
		length    = 5
	)

	b := make([]byte, length)
	for i := range b {
		b[i] = idCharset[rand.Intn(len(idCharset))]
	}
	return string(b)
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	log.Printf("get session %q", session)

	s, err := h.store.Load(session)
	if err != nil {
		showStoreError(w, err)
		return
	}

	if err := showJSON(w, s.Data); err != nil {
		return
	}
}

func (h *Handler) startVote(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["session"]

	log.Printf("start vote %q", id)

	_, err := store.ReadModifyWrite(id, h.store, func(s *domain.Session) (*domain.Session, error) {
		s.Open = true
		s.Votes = map[string]string{}

		return s, nil
	})

	switch err {
	case nil:
		w.WriteHeader(http.StatusAccepted)
	case store.ErrVersionMismatch:
		showError(w, http.StatusInternalServerError, "locking error, try again later", nil)
		return
	default:
		showStoreError(w, err)
		return
	}

	h.emitVoteEnabled(id)

}

func (h *Handler) stopVote(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["session"]

	log.Printf("stop vote %q", id)

	_, err := store.ReadModifyWrite(id, h.store, func(s *domain.Session) (*domain.Session, error) {
		s.Open = false

		return s, nil
	})

	switch err {
	case nil:
		w.WriteHeader(http.StatusAccepted)
	case store.ErrVersionMismatch:
		showError(w, http.StatusInternalServerError, "locking error, try again later", nil)
		return
	default:
		showStoreError(w, err)
		return
	}

	h.emitVoteDisabled(id)
}

func (h *Handler) resetVote(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["session"]

	log.Printf("reset vote %q", id)

	saved, err := store.ReadModifyWrite(id, h.store, func(s *domain.Session) (*domain.Session, error) {
		s.Open = false
		s.Votes = map[string]string{}

		return s, nil
	})

	switch err {
	case nil:
		w.WriteHeader(http.StatusAccepted)
	case store.ErrVersionMismatch:
		showError(w, http.StatusInternalServerError, "locking error, try again later", nil)
		return
	default:
		showStoreError(w, err)
		return
	}

	h.emitReset(id)
	h.emitVote(id, saved.Votes)

}

func (h *Handler) kickParticipant(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["session"]

	var name string
	if err := readContent(w, r, &name); err != nil {
		return
	}

	log.Printf("kick participant %q %q", id, name)

	saved, err := store.ReadModifyWrite(id, h.store, func(s *domain.Session) (*domain.Session, error) {
		idx := -1
		for i, p := range s.Participants {
			if p == name {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, errInvalidParticipant
		}
		s.Participants = append(s.Participants[:idx], s.Participants[idx+1:]...)

		return s, nil
	})

	switch err {
	case nil:
		w.WriteHeader(http.StatusNoContent)
	case errInvalidParticipant:
		showError(w, http.StatusBadRequest, "not a participant", nil)
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
