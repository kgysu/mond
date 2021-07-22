package mond

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

var UNHEALTHY = HealthCheck{
	Status:    "DOWN",
	Timestamp: 0,
}

var HEALTHY = HealthCheck{
	Status:    "UP",
	Timestamp: 1,
}

type HealthCheck struct {
	Status string `json:"status"`
	Timestamp int64 `json:"timestamp"`
}

func NewHealthCheck(rdr io.Reader) (*HealthCheck, error) {
	check := new(HealthCheck)
	err := json.NewDecoder(rdr).Decode(&check)

	if err != nil {
		err = fmt.Errorf("problem parsing apps, %v", err)
	}

	return check, err
}

func (h *HealthCheck) GetFormattedTime() string {
	return time.Unix(h.Timestamp, 0).Format("02.01.2006 15:04:05")
}
