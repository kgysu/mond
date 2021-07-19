package mond

type AccessLog struct {
	Timestamp int64  `json:"timestamp"`
	Ip        string `json:"ip"`
	Path      string `json:"path"`
	RemoteIp  string `json:"remoteIp"`
	Status    string `json:"status"`
	Raw       string `json:"raw"`
}

type AccessLogs []AccessLog
