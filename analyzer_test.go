package main

import (
	"reflect"
	"testing"
)

func TestAnalyzeLogStatement(t *testing.T) {
	t.Helper()
	rawLog := "10.129.38.1 - - [02/Jul/2021:22:50:59 +0200] \"GET /futures HTTP/1.1\" 200 7280 \"-\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36\" \"92.104.237.155\""

	got := ParseRawLog(rawLog)
	want := &LogEntry{
		ip: "10.129.38.1",
		timestamp: 1625259059000000000,
		status: "200",
		path: "/futures",
		method: "GET",
		http: "HTTP/1.1",
		xForwardedFor: "92.104.237.155",
		raw: rawLog,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
