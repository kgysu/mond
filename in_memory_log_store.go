package main

import (
	"time"
)

func NewInMemoryLogStore() *InMemoryLogStore {
	return &InMemoryLogStore{
		map[string][]*LogEntry{},
		LogStatistic{
			map[string]int{},
			map[int]int{},
		},
	}
}

type LogEntry struct {
	timestamp     int64
	path          string
	method        string
	http          string
	ip            string
	status        string
	xForwardedFor string
	raw           string
}

type LogStatistic struct {
	countPerIp  map[string]int
	countPerDay map[int]int
}

type InMemoryLogStore struct {
	store map[string][]*LogEntry
	stats LogStatistic
}

func (i *InMemoryLogStore) RecordLog(name string, value *LogEntry) {
	i.store[name] = append(i.store[name], value)
	i.stats.countPerIp[name]++
	i.stats.countPerDay[time.Now().YearDay()]++
}

func (i *InMemoryLogStore) GetRawLogs(name string) []string {
	var logs []string
	for _, log := range i.store[name] {
		logs = append(logs, log.raw)
	}
	return logs
}

func (i *InMemoryLogStore) GetApps() []string {
	var apps []string
	for k, _ := range i.store {
		apps = append(apps, k)
	}
	return apps
}
