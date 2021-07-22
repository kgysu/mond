package mond

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const HomePath = "/"
const AssetsPath = "/asset/"
const ApiAppsPath = "/apps/"
const ApiDashboardPath = "/dashboard/"
const DashboardAssetsPath = "/dashboard/asset/"
const DashboardRawLogsPath = "/dashboard/rawlogs/"
const ApiAccessLogsPath = "/logs/"
const ApiRawLogsPath = "/rawlogs/"
const ApiHealthPath = "/health/"

type AccessLogStore interface {
	GetAppNames() []string
	GetApps() Apps
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

type SecurityUserInfo struct {
	Username string
	Password string
}

const jsonContentType = "application/json"
const textContentType = "text/plain"

func NewApiServer(store AccessLogStore, info SecurityUserInfo) *ApiServer {
	s := new(ApiServer)
	s.store = store

	router := http.NewServeMux()
	router.Handle(ApiAppsPath, http.HandlerFunc(s.appsHandler))
	router.Handle(ApiDashboardPath, http.HandlerFunc(basicAuth(s.rootHandler, info)))
	router.Handle(DashboardRawLogsPath, http.HandlerFunc(basicAuth(s.rawLogsHandler, info)))
	router.Handle(ApiAccessLogsPath, http.HandlerFunc(s.logsHandler))
	router.Handle(ApiRawLogsPath, http.HandlerFunc(s.rawLogsHandler))
	router.Handle(ApiHealthPath, http.HandlerFunc(s.healthHandler))
	// assets
	fs := http.FileServer(http.Dir("asset/"))
	router.Handle(AssetsPath, http.StripPrefix(AssetsPath, fs))
	router.Handle(DashboardAssetsPath, http.StripPrefix(DashboardAssetsPath, fs))
	// root
	//router.Handle(HomePath, http.FileServer(http.Dir("./html")))
	router.Handle(HomePath, http.HandlerFunc(basicAuth(s.rootHandler, info)))

	s.Handler = router
	return s
}

func (s *ApiServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	apps := s.store.GetApps()
	if apps == nil || len(apps) < 1 {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	indexTempl := template.Must(template.ParseFiles("html/index.html"))
	err := indexTempl.Execute(w, apps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func (s *ApiServer) rawLogsHandler(w http.ResponseWriter, r *http.Request) {
	appName := strings.ToLower(strings.TrimPrefix(r.URL.Path, ApiRawLogsPath))
	switch r.Method {
	case http.MethodGet:
		s.showRawLogs(w, appName)
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

func (s *ApiServer) appsHandler(w http.ResponseWriter, r *http.Request) {
	apps := s.store.GetApps()
	if apps == nil || len(apps) < 1 {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(apps)
}

func (s *ApiServer) assetsHandler(writer http.ResponseWriter, request *http.Request) {
	fs := http.FileServer(http.Dir("assets/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}

func (s *ApiServer) showLogs(w http.ResponseWriter, name string) {
	logs := s.store.GetAccessLogs(name)

	if len(logs) < 1 {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Set("content-type", jsonContentType)
	json.NewEncoder(w).Encode(&logs)
}

func (s *ApiServer) showRawLogs(w http.ResponseWriter, name string) {
	logs := s.store.GetAccessLogs(name)

	if len(logs) < 1 {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Set("content-type", textContentType)
	for _, l := range logs {
		_, err := fmt.Fprintf(w, "%s\n", l.Raw)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

}

func (s *ApiServer) processLog(w http.ResponseWriter, name string, body io.ReadCloser) {
	bodyContent, err := ioutil.ReadAll(body)
	if err != nil {
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
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	s.store.RecordHealth(name, *parsedCheck)
	w.WriteHeader(http.StatusAccepted)
}

type handler func(w http.ResponseWriter, r *http.Request)

func basicAuth(pass handler, securityInfo SecurityUserInfo) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", "Basic realm=localhost") // TODO define realm
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		if u != securityInfo.Username || p != securityInfo.Password {
			w.Header().Set("WWW-Authenticate", "Basic realm=localhost")
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		pass(w, r)
	}
}
