package mond

import (
	"encoding/json"
	"fmt"
	"io"
)

var UNHEALTHY = HealthCheck{
	Status:    "DOWN",
	Timestamp: 0,
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
