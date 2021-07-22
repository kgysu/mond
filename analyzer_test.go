package mond

import (
	"testing"
)

func TestAnalyzeLogStatement(t *testing.T) {

	t.Run("parse valid sample log", func(t *testing.T) {
		rawLog := "10.129.38.1 - - [02/Jul/2021:22:50:59 +0200] \"GET /futures HTTP/1.1\" 200 7280 \"-\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36\" \"92.104.237.155\""

		got := ParseRawLog(rawLog)
		want := AccessLog{
			Ip:        "10.129.38.1",
			Timestamp: 1625259059,
			Status:    "200",
			Path:      "/futures",
			//method: "GET",
			//http: "HTTP/1.1",
			RemoteIp: "92.104.237.155",
			Raw:      rawLog,
		}

		assertAccessLogEquals(t, got, want)
	})

	t.Run("parse unusual log", func(t *testing.T) {
		rawLog := "SampleLog"

		got := ParseRawLog(rawLog)
		want := AccessLog{
			Ip:        "",
			Timestamp: 0,
			Status:    "",
			Path:      "",
			RemoteIp:  "",
			Raw:       rawLog,
		}

		assertAccessLogEquals(t, got, want)
	})

	t.Run("parse empty log", func(t *testing.T) {
		rawLog := ""

		got := ParseRawLog(rawLog)
		want := AccessLog{
			Ip:        "",
			Timestamp: 0,
			Status:    "",
			Path:      "",
			RemoteIp:  "",
			Raw:       rawLog,
		}

		assertAccessLogEquals(t, got, want)
	})
}

func assertAccessLogEquals(t testing.TB, got, want AccessLog) {
	t.Helper()
	if got.Ip != want.Ip {
		t.Errorf("got ip %v want %v", got.Ip, want.Ip)
	}
	if got.Path != want.Path {
		t.Errorf("got path %v want %v", got.Path, want.Path)
	}
	if got.RemoteIp != want.RemoteIp {
		t.Errorf("got remoteIp %v want %v", got.RemoteIp, want.RemoteIp)
	}
	if got.Status != want.Status {
		t.Errorf("got status %v want %v", got.Status, want.Status)
	}
	if got.Timestamp != want.Timestamp {
		t.Errorf("got timestamp %v want %v", got.Timestamp, want.Timestamp)
	}
	if got.Raw != want.Raw {
		t.Errorf("got raw %v want %v", got.Raw, want.Raw)
	}
}
