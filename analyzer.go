package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func ParseRawLog(raw string) *LogEntry {
	logEntry := new(LogEntry)
	logEntry.ip = findIp(raw)
	logEntry.timestamp = findTimeAndParse(raw)
	logEntry.status = findStatus(raw)
	path, method, http := findPathMethodHttp(raw)
	logEntry.path = path
	logEntry.method = method
	logEntry.http = http
	logEntry.xForwardedFor = findxForwardedFor(raw)
	logEntry.raw = raw
	return logEntry
}

func findIp(raw string) string {
	remoteReg := regexp.MustCompile(`^\d{1,3}\.\d{1,3}.\d{1,3}.\d{1,3}`)
	return remoteReg.FindString(raw)
}

func findTimeAndParse(raw string) int64 {
	timeReg := regexp.MustCompile(`\d{2}/.{2,3}/\d{4}:\d{2}:\d{2}:\d{2}\s.{5}`)
	timeString := timeReg.FindString(raw)
	if timeString != "" {
		t, err := time.Parse("02/Jan/2006:15:04:05 -0700", timeString)
		if err != nil {
			fmt.Printf("cannot parse time %q caused by %v", timeString, err)
			return 0
		} else {
			return t.UnixNano()
		}
	}
	return -1
}

func findStatus(raw string) string {
	statusReg := regexp.MustCompile(`\s\d{3}\s`)
	status := statusReg.FindString(raw)
	return strings.TrimSpace(status)
}

func findPathMethodHttp(raw string) (string, string, string) {
	pathReg := regexp.MustCompile(`["].*\sHTTP/\d\.\d["]`)
	rawPath := pathReg.FindString(raw)
	rawPath = strings.Trim(rawPath, "\"")

	path := ""
	method := ""
	http := ""

	splittedPath := strings.SplitN(rawPath, " ", 3)
	if len(splittedPath) > 0 {
		method = splittedPath[0]
	}
	if len(splittedPath) > 1 {
		path = splittedPath[1]
	}
	if len(splittedPath) > 2 {
		http = splittedPath[2]
	}
	return path, method, http
}

func findxForwardedFor(raw string) string {
	xIpReg := regexp.MustCompile(`["]\d{1,3}\.\d{1,3}.\d{1,3}.\d{1,3}["]`)
	xIp := xIpReg.FindString(raw)
	return strings.Trim(xIp, "\"")
}
