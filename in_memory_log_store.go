package main

import "sync"

func NewInMemoryLogStore() *InMemoryLogStore {
	return &InMemoryLogStore{
		map[string][]string{},
		sync.RWMutex{},
	}
}

type InMemoryLogStore struct {
	store map[string][]string
	// A mutex is used to synchronize read/write access to the map
	lock sync.RWMutex
}

func (i *InMemoryLogStore) RecordLog(name string, value string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.store[name] = append(i.store[name], value)
}

func (i *InMemoryLogStore) GetLogs(name string) []string {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.store[name]
}

func (i *InMemoryLogStore) GetApps() []string {
	i.lock.RLock()
	defer i.lock.RUnlock()
	var apps []string
	for k,_ := range i.store {
		apps = append(apps, k)
	}
	return apps
}
