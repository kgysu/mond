package mond

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CheckWebsite returns true if the URL returns a 200 status code, false otherwise.
func CheckWebsite(url string) HealthCheck {
	response, err := http.Head(url)
	if err != nil {
		return HealthCheck{
			Status:    "DOWN",
			Timestamp: time.Now().UnixNano(),
		}
	}

	if response.StatusCode != http.StatusOK {
		return HealthCheck{
			Status:    response.Status,
			Timestamp: time.Now().UnixNano(),
		}
	}

	return HealthCheck{
		Status:    "UP",
		Timestamp: time.Now().UnixNano(),
	}
}

// WebsiteChecker checks a url, returning a bool.
type WebsiteChecker func(string) HealthCheck
type result struct {
	string
	check HealthCheck
}

// CheckWebsites takes a WebsiteChecker and a slice of urls and returns  a map.
// of urls to the result of checking each url with the WebsiteChecker function.
func CheckWebsites(wc WebsiteChecker, urls []string) map[string]HealthCheck {
	results := make(map[string]HealthCheck)
	resultChannel := make(chan result)

	for _, url := range urls {
		go func(u string) {
			resultChannel <- result{u, wc(u)}
		}(url)
	}

	for i := 0; i < len(urls); i++ {
		r := <-resultChannel
		results[r.string] = r.check
	}

	return results
}

type WebsiteHealthReporter func(string, string) (*http.Response, error)


func Report(url string, content string) (*http.Response, error) {
	resp, err := http.Post(url, jsonContentType, strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("could not report: %v \n", err)
	}
	return resp, nil
}

func ReportHealthCheck(reporter WebsiteHealthReporter, url string, check HealthCheck) (int, error) {
	healthJson, err := json.Marshal(check)
	if err != nil {
		return 0, fmt.Errorf("cannot marshal %v", check)
	}
	reportResponse, err := reporter(url, string(healthJson))
	if err != nil {
		return 0, err
	}

	return reportResponse.StatusCode, nil
}
