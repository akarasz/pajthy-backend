package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/akarasz/pajthy-backend/handler"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	log.Fatal(http.ListenAndServe(":8000", handler.NewPajthyRouter()))
}
