package mond

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"
)

// Apps

type Apps []App

func (a Apps) Find(name string) *App {
	for i, v := range a {
		if v.Name == name {
			return &a[i]
		}
	}
	return nil
}

func NewApps(rdr io.Reader) (Apps, error) {
	var apps Apps
	err := json.NewDecoder(rdr).Decode(&apps)

	if err != nil {
		err = fmt.Errorf("problem parsing apps, %v", err)
	}

	return apps, err
}

// App

type App struct {
	Name   string      `json:"app"`
	Health HealthCheck `json:"health"`
	Logs   AccessLogs  `json:"logs"`
}

func (a *App) GetLogsSorted() AccessLogs {
	sort.Slice(a.Logs, func(i, j int) bool {
		return a.Logs[i].Unix > a.Logs[j].Unix
	})
	return a.Logs
}

func (a *App) GetLogCountPerDay() map[string]int {
	var logsPerDay map[string]int
	logsPerDay = map[string]int{}
	for _, l := range a.Logs {
		logTime := time.Unix(l.Unix, 0)
		y, m, d := logTime.Date()
		date := fmt.Sprintf("%d-%d-%d", y, m, d)
		logsPerDay[date]++
	}
	keys := make([]string, 0, len(logsPerDay))
	for k := range logsPerDay {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return logsPerDay
}

func (a *App) GetIpStatsSorted() []IpStat {
	ipStats := IpStats{}
	for _, l := range a.Logs {
		ipStats.Add(l)
	}
	return ipStats.Sorted()
}

type IpStat struct {
	Ip    string
	Count int
	Paths map[string]string
}

type IpStats struct {
	stats []IpStat
}

func (is *IpStats) Add(log AccessLog) {
	pathsKey := log.Path
	if len(pathsKey) > 19 {
		pathsKey = log.Path[:20]
	}
	stat := is.Find(log.RemoteIp)
	if stat == nil {
		is.stats = append(is.stats, IpStat{
			Ip:    log.RemoteIp,
			Count: 1,
			Paths: map[string]string{pathsKey: log.Path},
		})
	} else {
		stat.Count++
		stat.Paths[pathsKey] = log.Path
	}
}

func (is *IpStats) Find(ip string) *IpStat {
	for i, v := range is.stats {
		if v.Ip == ip {
			return &is.stats[i]
		}
	}
	return nil
}

func (is *IpStats) Sorted() []IpStat {
	sort.Slice(is.stats, func(i, j int) bool {
		return is.stats[i].Count > is.stats[j].Count
	})
	return is.stats
}
