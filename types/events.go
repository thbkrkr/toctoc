package types

import (
	"encoding/json"
	"time"
)

const (
	StatusOK = "OK"
	StatusKO = "KO"
)

type Event struct {
	ID        string
	Status    string
	Timestamp time.Time
	Value     map[string]interface{}
}

func (e Event) GetHost() string {
	return e.Value["Host"].(string)
}

func (e Event) GetService() string {
	return e.Value["Service"].(string)
}

func (e Event) GetStatus() string {
	return e.Value["Status"].(string)
}

func (e Event) ToBytes() ([]byte, error) {
	bytes, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
