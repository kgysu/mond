package mond

import (
	"encoding/json"
	"fmt"
	"io"
)

type Apps []AppAccessLogs

func (a Apps) Find(name string) *AppAccessLogs {
	for i, v := range a {
		if v.App == name {
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
