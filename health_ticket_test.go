package mond

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

const invalidUrl = "invalid"

func mockWebsiteChecker(url string) HealthCheck {
	if url == invalidUrl {
		return UNHEALTHY
	}
	return HEALTHY
}

func TestCheckWebsites(t *testing.T) {
	websites := []string{
		"http://localhost:5000",
		invalidUrl,
	}

	want := map[string]HealthCheck{
		"http://localhost:5000": HEALTHY,
		invalidUrl:              UNHEALTHY,
	}

	got := CheckWebsites(mockWebsiteChecker, websites)

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Wanted %v, got %v", want, got)
	}
}


func mockHealthReporter (url string, content string) (*http.Response, error) {
	if url == "valid" {
		var check HealthCheck
		err := json.Unmarshal([]byte(content), &check)
		if err != nil {
			return &http.Response{StatusCode: http.StatusBadRequest}, nil
		}
		return &http.Response{StatusCode: http.StatusAccepted}, nil
	}
	if url == "invalid" {
		return nil, fmt.Errorf("invalid url")
	}
	return &http.Response{StatusCode: http.StatusNotFound}, nil
}

func TestReportHealthCheck(t *testing.T) {

	t.Run("reports HEALTHY successfully", func(t *testing.T) {
		want := &http.Response{StatusCode: http.StatusAccepted}
		got, err := ReportHealthCheck(mockHealthReporter, "valid", HEALTHY)

		assertNoError(t, err)
		assertStatus(t, got, want.StatusCode)
	})

	t.Run("reports UNHEALTHY successfully", func(t *testing.T) {
		want := &http.Response{StatusCode: http.StatusAccepted}
		got, err := ReportHealthCheck(mockHealthReporter, "valid", UNHEALTHY)

		assertNoError(t, err)
		assertStatus(t, got, want.StatusCode)
	})

	t.Run("reports HealthCheck on invalid url, should return bad request", func(t *testing.T) {
		_, err := ReportHealthCheck(mockHealthReporter, "invalid", HealthCheck{})

		if err == nil {
			t.Fatalf("error expected but got %v instead", err)
		}
	})
}
