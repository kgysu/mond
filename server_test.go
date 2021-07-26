package mond

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
var testInfo = SecurityUserInfo{
	Username: "test",
	Password: "test",
}

func TestGETHome(t *testing.T) {

	t.Run("returns OK on HomePath", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, HomePath, nil)
		response := httptest.NewRecorder()
		server := NewApiServer(&StubLogStore{}, testInfo)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns Dashboard on valid credentials", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, DashboardPath, nil)
		request.SetBasicAuth(testInfo.Username, testInfo.Password)
		response := httptest.NewRecorder()
		emptyStore := StubLogStore{[]App{{
			Name:   "Test",
			Health: HealthCheck{},
			Logs:   nil,
		}}}
		server := NewApiServer(&emptyStore, testInfo)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns 401 Unauthorized on invalid credentials", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, DashboardPath, nil)
		request.SetBasicAuth("some", "other")
		response := httptest.NewRecorder()
		emptyStore := StubLogStore{[]App{}}
		server := NewApiServer(&emptyStore, testInfo)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("returns StatusNotFound on valid credentials and empty store", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, DashboardPath, nil)
		request.SetBasicAuth(testInfo.Username, testInfo.Password)
		response := httptest.NewRecorder()
		emptyStore := StubLogStore{[]App{}}
		server := NewApiServer(&emptyStore, testInfo)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})
}

func TestGETLogsAndHealth(t *testing.T) {
	wantedLogsAppA := AccessLogs{
		{Raw: "log a1"},
		{Raw: "log a2"},
	}
	wantedLogsAppB := AccessLogs{
		{Raw: "log b1"},
		{Raw: "log b2"},
	}
	store := StubLogStore{
		[]App{
			{"appa", HEALTHY, wantedLogsAppA},
			{"appb", UNHEALTHY, wantedLogsAppB},
		},
	}
	server := NewApiServer(&store, testInfo)

	t.Run("returns Logs of App A", func(t *testing.T) {
		request := newGetLogsRequest("AppA")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := decodeBodyToAccessLogs(t, response.Body)

		assertStatus(t, response.Code, http.StatusOK)
		assertAccessLogsEquals(t, got, wantedLogsAppA)
		assertContentType(t, response, jsonContentType)
	})

	t.Run("returns RawLogs of App A", func(t *testing.T) {
		request := newGetRawLogsRequest("AppA")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Body.String()
		want := "log a1\nlog a2\n"
		assertStatus(t, response.Code, http.StatusOK)
		if got != want {
			t.Errorf("did not get correct rawlogs, got %q, want %q", got, want)
		}
		assertContentType(t, response, textContentType)
	})

	t.Run("returns Health of AppA", func(t *testing.T) {
		request := newGetHealthRequest("AppA")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := decodeBodyToHealth(t, response.Body)

		assertStatus(t, response.Code, http.StatusOK)
		assertHealthEquals(t, got, HEALTHY)
		assertContentType(t, response, jsonContentType)
	})

	t.Run("returns Logs of App B", func(t *testing.T) {
		request := newGetLogsRequest("AppB")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := decodeBodyToAccessLogs(t, response.Body)

		assertStatus(t, response.Code, http.StatusOK)
		assertAccessLogsEquals(t, got, wantedLogsAppB)
		assertContentType(t, response, jsonContentType)
	})

	t.Run("returns Health of AppB", func(t *testing.T) {
		request := newGetHealthRequest("AppB")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := decodeBodyToHealth(t, response.Body)

		assertStatus(t, response.Code, http.StatusOK)
		assertHealthEquals(t, got, UNHEALTHY)
		assertContentType(t, response, jsonContentType)
	})
}

func TestStoreLogs(t *testing.T) {

	t.Run("it records and analyzes logs on POST", func(t *testing.T) {
		store := StubLogStore{}
		server := NewApiServer(&store, testInfo)
		app := "App1"
		request := newPostLogRequest(app)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusAccepted)

		if len(store.AppAccessLogs) != 1 {
			t.Fatalf("got %d calls to RecordLog want %d", len(store.AppAccessLogs), 1)
		}
		if store.AppAccessLogs[0].Logs[0].Raw != SampleLogA1 {
			t.Errorf("did not store correct log got %q want %q", store.AppAccessLogs[0].Logs[0].Raw, SampleLogA1)
		}
	})

	t.Run("it records health on POST", func(t *testing.T) {
		store := StubLogStore{}
		server := NewApiServer(&store, testInfo)
		app := "AppA"
		request := newPostHealthRequest(app)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusAccepted)

		if len(store.AppAccessLogs) != 1 {
			t.Fatalf("got %d calls to RecordLog want %d", len(store.AppAccessLogs), 1)
		}
		assertHealthEquals(t, store.AppAccessLogs[0].Health, HEALTHY)
	})
}

func decodeBodyToStringArray(t testing.TB, body io.Reader) (logs []string) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&logs)

	if err != nil {
		t.Fatalf("Unable to parse response from server '%v' into string array, '%v'", body, err)
	}
	return
}

func decodeBodyToApps(t testing.TB, body io.Reader) (apps Apps) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&apps)

	if err != nil {
		t.Fatalf("Unable to parse response from server '%v' into string array, '%v'", body, err)
	}
	return
}

func decodeBodyToAccessLogs(t testing.TB, body io.Reader) (logs AccessLogs) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&logs)

	if err != nil {
		t.Fatalf("Unable to parse response from server '%v' into AccessLogs, '%v'", body, err)
	}
	return
}

func decodeBodyToHealth(t testing.TB, body io.Reader) (check HealthCheck) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&check)

	if err != nil {
		t.Fatalf("Unable to parse response from server '%v' into HealthCheck, '%v'", body, err)
	}
	return
}

func newPostLogRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf(ApiAccessLogsPath+"%s", name), bytes.NewReader([]byte(SampleLogA1)))
	return req
}

func newPostHealthRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf(ApiHealthPath+"%s", name), bytes.NewReader([]byte(`{"status":"UP","timestamp":1}`)))
	return req
}

func newGetLogsRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf(ApiAccessLogsPath+"%s", name), nil)
	return req
}

func newGetRawLogsRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf(ApiRawLogsPath+"%s", name), nil)
	return req
}

func newGetHealthRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf(ApiHealthPath+"%s", name), nil)
	return req
}

func assertStringArray(t testing.TB, got, want []string) {
	t.Helper()
	if len(want) == 0 && len(got) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertApps(t testing.TB, got, want Apps) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertHealthEquals(t testing.TB, got, want HealthCheck) {
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
