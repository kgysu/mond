package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecordingLogsAndRetrievingThem(t *testing.T) {
	store := NewInMemoryLogStore()
	server := NewApiServer(store)
	app := "AppA"

	server.ServeHTTP(httptest.NewRecorder(), newPostLogRequest(app))
	server.ServeHTTP(httptest.NewRecorder(), newPostLogRequest(app))
	server.ServeHTTP(httptest.NewRecorder(), newPostLogRequest(app))
	server.ServeHTTP(httptest.NewRecorder(), newPostLogRequest("Other"))

	t.Run("get Apps", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		got := decodeBodytoStringArray(t, response.Body)
		want := []string{
			app,
			"Other",
		}
		assertStringArray(t, got, want)
	})

	t.Run("get logs", func(t *testing.T) {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, newGetLogsRequest(app))
		assertStatus(t, response.Code, http.StatusOK)

		got := decodeBodytoStringArray(t, response.Body)
		want := []string{
			SampleLogA1,
			SampleLogA1,
			SampleLogA1,
		}
		assertStringArray(t, got, want)
	})
}
