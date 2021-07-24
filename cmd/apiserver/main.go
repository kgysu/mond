package main

import (
	"fmt"
	"log"
	mond "mond-api"
	"net/http"
	"os"
)

const dbFileNameEnv = "MOND_DB_FILE_NAME"
const usernameEnv = "MOND_USERNAME"
const passwordEnv = "MOND_PW"
const addrEnv = "MOND_SERVE_ADDR"
const defaultDbFileName = "apps.db.json"
const defaultUser = "test"
const defaultPassword = "1234"
const defaultAddr = ":8080"

func main() {
	store, closeFile, err := mond.FileSystemAppsStoreFromFile(dbFileNameFromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile()

	server := mond.NewApiServer(store, checkEnvSecurityInfo())
	log.Fatal(http.ListenAndServe(addrFromEnv(), server))
}

func dbFileNameFromEnv() string {
	dbFileName := os.Getenv(dbFileNameEnv)
	if dbFileName == "" {
		dbFileName = defaultDbFileName
	}
	return dbFileName
}

func addrFromEnv() string {
	addr := os.Getenv(addrEnv)
	if addr == "" {
		addr = defaultAddr
	}
	fmt.Println("Run on ", addr)
	return addr
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
