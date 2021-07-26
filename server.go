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
const DashboardPath = "/dashboard/"
const DashboardAssetsPath = "/dashboard/asset/"
const DashboardRawLogsPath = "/dashboard/rawlogs/"
const DashboardLogsPath = "/dashboard/logs/"
const DashboardStatsPath = "/dashboard/stats/"
const DashboardReqsPath = "/dashboard/reqs/"
const ApiAccessLogsPath = "/logs/"
const ApiRawLogsPath = "/rawlogs/"
const ApiHealthPath = "/health/"

type AccessLogStore interface {
	GetAppNames() []string
	GetApps() Apps
	GetApp(name string) *App
	GetAccessLogs(name string) AccessLogs
	RecordAccessLog(name string, value AccessLog)
	GetHealth(name string) HealthCheck
	RecordHealth(name string, check HealthCheck)
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
	// Dashboard
	router.Handle(DashboardPath, http.HandlerFunc(basicAuth(s.dashboardHandler, info)))
	router.Handle(DashboardRawLogsPath, http.HandlerFunc(basicAuth(s.rawLogsHandler, info)))
	router.Handle(DashboardLogsPath, http.HandlerFunc(basicAuth(s.dashboardLogsHandler, info)))
	router.Handle(DashboardStatsPath, http.HandlerFunc(basicAuth(s.statsHandler, info)))
	router.Handle(DashboardReqsPath, http.HandlerFunc(basicAuth(s.reqsHandler, info)))
	fs := http.FileServer(http.Dir("asset/"))
	router.Handle(DashboardAssetsPath, http.StripPrefix(DashboardAssetsPath, fs))

	// API
	//router.Handle(ApiAppsPath, http.HandlerFunc(s.appsHandler))
	// TODO check to delete
	router.Handle(ApiAccessLogsPath, http.HandlerFunc(s.logsHandler))
	router.Handle(ApiRawLogsPath, http.HandlerFunc(s.rawLogsHandler))
	router.Handle(ApiHealthPath, http.HandlerFunc(s.healthHandler))

	// Root
	//router.Handle(HomePath, http.FileServer(http.Dir("./html")))
	router.Handle(HomePath, http.HandlerFunc(s.rootHandler))

	s.Handler = router
	return s
}

func (s *ApiServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *ApiServer) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
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

func (s *ApiServer) statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	appName := strings.ToLower(strings.TrimPrefix(r.URL.Path, DashboardStatsPath))
	app := s.store.GetApp(appName)
	if app == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	stats := app.GetIpStatsSorted()
	indexTempl := template.Must(template.ParseFiles("html/ipstats.html"))
	err := indexTempl.Execute(w, stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *ApiServer) reqsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	appName := strings.ToLower(strings.TrimPrefix(r.URL.Path, DashboardReqsPath))
	app := s.store.GetApp(appName)
	if app == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	stats := app.GetLogCountPerDay()
	indexTempl := template.Must(template.ParseFiles("html/reqsPerDay.html"))
	err := indexTempl.Execute(w, stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *ApiServer) dashboardLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	appName := strings.ToLower(strings.TrimPrefix(r.URL.Path, DashboardLogsPath))
	app := s.store.GetApp(appName)
	if app == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	indexTempl := template.Must(template.ParseFiles("html/logs.html"))
	err := indexTempl.Execute(w, app.GetLogsSorted())
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
	appName := strings.ToLower(r.URL.Path)
	if strings.Contains(appName, ApiRawLogsPath) {
		appName = strings.TrimPrefix(appName, ApiRawLogsPath)
	}
	if strings.Contains(appName, DashboardRawLogsPath) {
		appName = strings.TrimPrefix(appName, DashboardRawLogsPath)
	}
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
