package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	event "github.com/akarasz/pajthy-backend/event/local"
	"github.com/akarasz/pajthy-backend/handler"
	store "github.com/akarasz/pajthy-backend/store/local"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	s := store.New()
	e := event.New()

	log.Fatal(http.ListenAndServe(":8000", handler.NewRouter(s, e)))
}
