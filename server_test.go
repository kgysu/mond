package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const SampleLogA1 = "Log a1"

type StubLogStore struct {
	logs map[string][]string
}

func (s *StubLogStore) GetLogs(name string) []string {
	logs := s.logs[name]
	return logs
}

func (s *StubLogStore) RecordLog(name string, value string) {
	s.logs[name] = append(s.logs[name], value)
}

func TestGETHome(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	store := StubLogStore{}
	server := NewApiServer(&store)

	t.Run("returns Home Page", func(t *testing.T) {
		server.ServeHTTP(response, request)

		assertResponseBody(t, response.Body.String(), "OK")
	})
}

func TestGETLogs(t *testing.T) {
	wantedLogsAppA := []string{
		"log a1",
		"log a2",
	}
	wantedLogsAppB := []string{
		"log b1",
		"log b2",
	}
	store := StubLogStore{
		map[string][]string{
			"AppA": wantedLogsAppA,
			"AppB": wantedLogsAppB,
		},
	}
	server := NewApiServer(&store)

	t.Run("returns Logs of App A", func(t *testing.T) {
		request := newGetLogsRequest("AppA")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := getLogsFromResponse(t, response.Body)

		assertStatus(t, response.Code, http.StatusOK)
		assertLogs(t, got, wantedLogsAppA)
		assertContentType(t, response, jsonContentType)
	})

	t.Run("returns Logs of App B", func(t *testing.T) {
		request := newGetLogsRequest("AppB")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := getLogsFromResponse(t, response.Body)

		assertStatus(t, response.Code, http.StatusOK)
		assertLogs(t, got, wantedLogsAppB)
		assertContentType(t, response, jsonContentType)
	})
}

func TestStoreLogs(t *testing.T) {
	store := StubLogStore{
		map[string][]string{},
	}
	server := NewApiServer(&store)

	t.Run("it records logs on POST", func(t *testing.T) {
		app := "App1"

		request := newPostLogRequest(app)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusAccepted)

		if len(store.logs) != 1 {
			t.Fatalf("got %d calls to RecordLog want %d", len(store.logs), 1)
		}

		if store.logs[app][0] != SampleLogA1 {
			t.Errorf("did not store correct log got %q want %q", store.logs[app][0], SampleLogA1)
		}
	})
}

func getLogsFromResponse(t testing.TB, body io.Reader) (logs []string) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&logs)

	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of Player, '%v'", body, err)
	}

	return
}

func newPostLogRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf(ApiLogsPath+"%s", name), bytes.NewReader([]byte(SampleLogA1)))
	return req
}

func newGetLogsRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf(ApiLogsPath+"%s", name), nil)
	return req
}

func assertLogs(t testing.TB, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertContentType(t testing.TB, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	if response.Result().Header.Get("content-type") != want {
		t.Errorf("response did not have content-type of %s, got %v", want, response.Result().Header)
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("response body is wrong, got %q want %q", got, want)
	}
}
