package mond

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileSystemAppsStore struct {
	database *json.Encoder
	apps     Apps
}

// NewFileSystemAppsStore creates a FileSystemAppsStore initialising the store if needed.
func NewFileSystemAppsStore(file *os.File) (*FileSystemAppsStore, error) {
	err := initialiseAppsDBFile(file)
	if err != nil {
		return nil, fmt.Errorf("problem initialising apps db file, %v", err)
	}

	apps, err := NewApps(file)
	if err != nil {
		return nil, fmt.Errorf("problem loading apps store from file %s, %v", file.Name(), err)
	}

	return &FileSystemAppsStore{
		database: json.NewEncoder(&tape{file}),
		apps:     apps,
	}, nil
}

// FileSystemAppsStoreFromFile creates a FileSystemAppsStore from the contents of a JSON file found at path.
func FileSystemAppsStoreFromFile(path string) (*FileSystemAppsStore, func(), error) {
	db, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return nil, nil, fmt.Errorf("problem opening %s, %v", path, err)
	}

	closeFunc := func() {
		db.Close()
	}

	store, err := NewFileSystemAppsStore(db)

	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("problem creating file system player store, %v ", err)
	}

	return store, closeFunc, nil
}

func initialiseAppsDBFile(file *os.File) error {
	file.Seek(0, 0)

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("problem getting file info from file %s, %v", file.Name(), err)
	}

	if info.Size() == 0 {
		file.Write([]byte("[]"))
		file.Seek(0, 0)
	}
	return nil
}

func (f *FileSystemAppsStore) GetAppNames() []string {
	var apps []string
	for _, v := range f.apps {
		apps = append(apps, v.App)
	}
	return apps
}

func (f *FileSystemAppsStore) GetApps() Apps {
	return f.apps
}

func (f *FileSystemAppsStore) GetAccessLogs(name string) AccessLogs {
	app := f.apps.Find(name)
	if app != nil {
		return app.Logs
	}
	return AccessLogs{}
}

func (f *FileSystemAppsStore) RecordAccessLog(name string, log AccessLog) {
	app := f.apps.Find(name)
	if app != nil {
		app.Logs = append(app.Logs, log)
	} else {
		f.apps = append(f.apps, AppAccessLogs{
			App:  name,
			Logs: AccessLogs{log},
		})
	}
	f.database.Encode(f.apps)
}

func (f *FileSystemAppsStore) GetHealth(name string) HealthCheck {
	app := f.apps.Find(name)
	if app != nil {
		return app.Health
	}
	return UNHEALTHY
}

func (f *FileSystemAppsStore) RecordHealth(name string, check HealthCheck) {
	app := f.apps.Find(name)
	if app != nil {
		app.Health = check
	} else {
		f.apps = append(f.apps, AppAccessLogs{
			App:    name,
			Health: check,
		})
	}
	f.database.Encode(f.apps)
}
