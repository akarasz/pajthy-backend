package handler

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/akarasz/pajthy-backend/domain"
)

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	log.Print("create session")

	var choices []string
	if err := parseContent(w, r, &choices); err != nil {
		return
	}

	id := generateID()
	s := domain.NewSession()
	s.Choices = choices

	if err := h.store.Create(id, s); err != nil {
		handleStoreError(w, err)
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

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	if err := sendJSON(w, s); err != nil {
		return
	}
}

func (h *Handler) startVote(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	log.Printf("start vote %q", session)

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	s.Open = true
	s.Votes = map[string]string{}

	if err := h.store.Update(session, s); err != nil {
		handleStoreError(w, err)
		return
	}

	h.emitVoteEnabled(session)

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) stopVote(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	log.Printf("stop vote %q", session)

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	s.Open = false

	if err := h.store.Update(session, s); err != nil {
		handleStoreError(w, err)
		return
	}

	h.emitVoteDisabled(session)

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) resetVote(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	log.Printf("reset vote %q", session)

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	s.Open = false
	s.Votes = map[string]string{}

	if err := h.store.Update(session, s); err != nil {
		handleStoreError(w, err)
		return
	}

	h.emitReset(session)
	h.emitVote(session, s.Votes)

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) kickParticipant(w http.ResponseWriter, r *http.Request) {
	session := mux.Vars(r)["session"]

	var name string
	if err := parseContent(w, r, &name); err != nil {
		return
	}

	log.Printf("kick participant %q %q", session, name)

	s, err := h.store.LockAndLoad(session)
	defer h.store.Unlock(session)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	idx := -1
	for i, p := range s.Participants {
		if p == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		http.Error(w, "not a participant", http.StatusBadRequest)
		return
	}
	s.Participants = append(s.Participants[:idx], s.Participants[idx+1:]...)

	if err := h.store.Update(session, s); err != nil {
		handleStoreError(w, err)
		return
	}

	h.emitParticipantsChange(session, s.Participants)

	w.WriteHeader(http.StatusNoContent)
}
