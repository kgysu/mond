package mond

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

var systemTZ = os.Getenv("TZ")

func ParseRawLog(raw string) AccessLog {
	accessLog := new(AccessLog)
	accessLog.Ip = findIp(raw)
	accessLog.Timestamp = findTimeAndParse(raw)
	accessLog.Unix = time.Now().Unix()
	accessLog.Status = findStatus(raw)
	path, _, _ := findPathMethodHttp(raw)
	accessLog.Path = path
	accessLog.RemoteIp = findxForwardedFor(raw)
	accessLog.Raw = raw
	return *accessLog
}

func findIp(raw string) string {
	remoteReg := regexp.MustCompile(`^\d{1,3}\.\d{1,3}.\d{1,3}.\d{1,3}`)
	return remoteReg.FindString(raw)
}

func findTimeAndParse(raw string) int64 {
	timeReg := regexp.MustCompile(`\d{2}/.{2,3}/\d{4}:\d{2}:\d{2}:\d{2}\s.{5}`)
	timeString := timeReg.FindString(raw)
	if timeString != "" {
		var t time.Time
		var err error
		l := getLocalTimeZone()
		if l == nil {
			t, err = time.Parse("02/Jan/2006:15:04:05 -0700", timeString)
		} else {
			t, err = time.ParseInLocation("02/Jan/2006:15:04:05 -0700", timeString, l)
		}
		if err != nil {
			fmt.Printf("cannot parse time %q caused by %v\n", timeString, err)
			return 0
		} else {
			return t.Unix()
		}
	}
	return 0
}

func getLocalTimeZone() *time.Location {
	l, err := time.LoadLocation(systemTZ)
	if err != nil {
		fmt.Printf("cannot load local timezone %q caused by %v\n", systemTZ, err)
		return nil
	}
	return l
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
