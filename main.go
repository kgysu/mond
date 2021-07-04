package main

import (
	"log"
	"net/http"
)

func main() {
	server := NewApiServer(NewInMemoryLogStore())
	log.Fatal(http.ListenAndServe(":5000", server))
}
