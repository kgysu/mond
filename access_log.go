package mond

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AccessLog struct {
	Timestamp int64  `json:"timestamp"`
	Unix      int64  `json:"unix"`
	Ip        string `json:"ip"`
	Path      string `json:"path"`
	RemoteIp  string `json:"remoteIp"`
	Status    string `json:"status"`
	Raw       string `json:"raw"`
}

type AccessLogs []AccessLog

func ReportRawLog(url string, content string) error {
	resp, err := http.Post(url, jsonContentType, strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("could not report: %v \n", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("got wrong response code, got %d want 202 \n", resp.StatusCode)
	}
	return nil
}

func (log *AccessLog) GetUnixFormatted() string {
	return time.Unix(log.Unix, 0).Format("02.01.2006 15:04:05")
}

func (log *AccessLog) GetTimestampFormatted() string {
	return time.Unix(log.Timestamp, 0).Format("02.01.2006 15:04:05")
}

func (log *AccessLog) GetRawIfNotAnalysed() string {
	if log.Ip != "" {
		return ""
	}
	return log.Raw
}
