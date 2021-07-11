package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const ApiLogsPath = "/logs/"
const ApiAnalyticsPath = "/analytics/"

type LogsStore interface {
	GetRawLogs(name string) []string
	RecordLog(name string, value *LogEntry)
	GetApps() []string
}


type ApiServer struct {
	store LogsStore
	http.Handler
}

const jsonContentType = "application/json"

func NewApiServer(store LogsStore) *ApiServer {
	s := new(ApiServer)

	s.store = store

	router := http.NewServeMux()
	router.Handle(ApiLogsPath, http.HandlerFunc(s.logsHandler))
	//router.Handle(ApiAnalyticsPath, http.HandlerFunc(s.analyticsHandler))
	router.Handle("/", http.HandlerFunc(s.rootHandler))

	s.Handler = router

	return s
}

func (s *ApiServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	apps := s.store.GetApps()

	if len(apps) < 1 {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(apps)
}

func (s *ApiServer) logsHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, ApiLogsPath)
	switch r.Method {
	case http.MethodPost:
		s.processLog(w, name, r.Body)
	case http.MethodGet:
		s.showLogs(w, name)
	}
}

func (s *ApiServer) analyticsHandler(w http.ResponseWriter, r *http.Request) {
	//name := strings.TrimPrefix(r.URL.Path, ApiLogsPath)
	//switch r.Method {
	//case http.MethodGet:
	//	//s.showAnalytics(w, name)
	//}
}

func (s *ApiServer) showLogs(w http.ResponseWriter, name string) {
	logs := s.store.GetRawLogs(name)

	if len(logs) < 1 {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(logs)
}

func (s *ApiServer) processLog(w http.ResponseWriter, name string, body io.ReadCloser) {
	bodyContent, err := ioutil.ReadAll(body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	s.store.RecordLog(name, ParseRawLog(string(bodyContent)))
	w.WriteHeader(http.StatusAccepted)
}
