package mond

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const HomePath = "/"
const ApiAccessLogsPath = "/logs/"
const ApiHealthPath = "/health/"

type AccessLogStore interface {
	GetAppNames() []string
	GetAccessLogs(name string) AccessLogs
	RecordAccessLog(name string, value AccessLog)
	GetHealth(name string) HealthCheck
	RecordHealth(name string, check HealthCheck)
}

type AppAccessLogs struct {
	App    string      `json:"app"`
	Health HealthCheck `json:"health"`
	Logs   AccessLogs  `json:"logs"`
}

type ApiServer struct {
	store AccessLogStore
	http.Handler
}

const jsonContentType = "application/json"

func NewApiServer(store AccessLogStore) *ApiServer {
	s := new(ApiServer)
	s.store = store

	router := http.NewServeMux()
	router.Handle(ApiAccessLogsPath, http.HandlerFunc(s.logsHandler))
	router.Handle(ApiHealthPath, http.HandlerFunc(s.healthHandler))
	router.Handle(HomePath, http.HandlerFunc(s.rootHandler))

	s.Handler = router
	return s
}

func (s *ApiServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	apps := s.store.GetAppNames()
	if apps == nil || len(apps) < 1 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "")
		return
	}

	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(apps)
}

func (s *ApiServer) logsHandler(w http.ResponseWriter, r *http.Request) {
	appName := strings.ToLower(strings.TrimPrefix(r.URL.Path, ApiAccessLogsPath))
	switch r.Method {
	case http.MethodPost:
		s.processLog(w, appName, r.Body)
	case http.MethodGet:
		s.showLogs(w, appName)
	}
}

func (s *ApiServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.ToLower(strings.TrimPrefix(r.URL.Path, ApiHealthPath))
	switch r.Method {
	case http.MethodPost:
		s.processHealth(w, name, r.Body)
	case http.MethodGet:
		s.showHealth(w, name)
	}
}

func (s *ApiServer) showLogs(w http.ResponseWriter, name string) {
	logs := s.store.GetAccessLogs(name)

	if len(logs) < 1 {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(&logs)
}

func (s *ApiServer) processLog(w http.ResponseWriter, name string, body io.ReadCloser) {
	bodyContent, err := ioutil.ReadAll(body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	s.store.RecordAccessLog(name, ParseRawLog(string(bodyContent)))
	w.WriteHeader(http.StatusAccepted)
}

func (s *ApiServer) showHealth(w http.ResponseWriter, name string) {
	health := s.store.GetHealth(name)
	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(&health)
}

func (s *ApiServer) processHealth(w http.ResponseWriter, name string, body io.ReadCloser) {
	parsedCheck, err := NewHealthCheck(body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	s.store.RecordHealth(name, *parsedCheck)
	w.WriteHeader(http.StatusAccepted)
}
