package main

import (
	"fmt"
	"log"
	mond "mond-api"
	"net/http"
	"os"
)

const dbFileName = "apps.db.json"
const usernameEnv = "MOND_USERNAME"
const passwordEnv = "MOND_PW"
const defaultUser = "test"
const defaultPassword = "1234"

func main() {
	store, closeFile, err := mond.FileSystemAppsStoreFromFile(dbFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile()

	server := mond.NewApiServer(store, checkEnvSecurityInfo())
	log.Fatal(http.ListenAndServe(":5000", server))
}

func checkEnvSecurityInfo() mond.SecurityUserInfo {
	username := os.Getenv(usernameEnv)
	if username == "" {
		username = defaultUser
		fmt.Println("WARN: using default user!")
	}
	pw := os.Getenv(passwordEnv)
	if pw == "" {
		pw = defaultPassword
		fmt.Println("WARN: using default password!")
	}
	return mond.SecurityUserInfo{Username: username, Password: pw}
}
