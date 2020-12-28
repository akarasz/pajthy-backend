package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/akarasz/pajthy-backend/event"
	"github.com/akarasz/pajthy-backend/handler"
	"github.com/akarasz/pajthy-backend/store"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	s := store.NewInternal()
	e := event.NewInternal()

	log.Fatal(http.ListenAndServe(":8000", handler.NewRouter(s, e)))
}
