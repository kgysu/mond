package main

import (
	"log"
	mond "mond-api"
	"net/http"
)

const dbFileName = "apps.db.json"

func main() {
	store, closeFile, err := mond.FileSystemAppsStoreFromFile(dbFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile()

	server := mond.NewApiServer(store)
	log.Fatal(http.ListenAndServe(":5000", server))
}
