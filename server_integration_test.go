package mond

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecordingLogsAndRetrievingThem(t *testing.T) {
	database, cleanDatabase := createTempFile(t, `[]`)
	defer cleanDatabase()
	store, err := NewFileSystemAppsStore(database)

	assertNoError(t, err)

	server := NewApiServer(store)
	app := "AppA"

	server.ServeHTTP(httptest.NewRecorder(), newPostLogRequest(app))
	server.ServeHTTP(httptest.NewRecorder(), newPostHealthRequest(app))
	server.ServeHTTP(httptest.NewRecorder(), newPostLogRequest(app))
	server.ServeHTTP(httptest.NewRecorder(), newPostLogRequest(app))
	server.ServeHTTP(httptest.NewRecorder(), newPostLogRequest("Other"))

	t.Run("get Apps", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		got := decodeBodyToStringArray(t, response.Body)
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

		got := decodeBodyToAccessLogs(t, response.Body)
		want := AccessLogs{
			{Raw: SampleLogA1},
			{Raw: SampleLogA1},
			{Raw: SampleLogA1},
		}
		assertAccessLogsEquals(t, got, want)
	})

	t.Run("get health", func(t *testing.T) {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, newGetHealthRequest(app))
		assertStatus(t, response.Code, http.StatusOK)

		got := decodeBodyToHealth(t, response.Body)
		want := HEALTHY
		assertHealthEquals(t, got, want)
	})
}
