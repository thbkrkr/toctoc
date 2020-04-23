package types

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	StatusOK = "OK"
	StatusKO = "KO"
)

type Event struct {
	TTL       float64
	ID        string
	Status    string
	Timestamp time.Time
	Value     map[string]interface{}
}

func (e Event) GetCheckTTL() float64 {
	return e.Value["CheckTTL"].(float64)
}

func (e Event) GetHost() string {
	if e.Value["Host"] == nil {
		return "undefined"
	}
	return e.Value["Host"].(string)
}

func (e Event) GetService() string {
	return e.Value["Service"].(string)
}

func (e Event) GetMessage() string {
	if e.Value["Message"] == nil {
		return "undefined"
	}
	return e.Value["Message"].(string)
}

func (e Event) IsKO() bool {
	if e.Status == StatusKO {
		return true
	}
	if e.TTL < 0 {
		return false
	}
	if time.Since(e.Timestamp).Seconds() > e.TTL {
		return true
	}
	return false
}

func (e Event) ToBytes() ([]byte, error) {
	bytes, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// parseEvent parses a generic JSON object in a structured event
// Host: Host || Node
// Service: Service
// Status: Status || State
func ParseEvent(defaultCheckTTL float64, obj map[string]interface{}) (Event, error) {
	host := obj["Host"]
	if host == nil {
		host = obj["Node"]
	}
	if host == nil {
		return Event{}, errors.New("Property 'Host' not found")
	}
	service := obj["Service"]
	if service == nil {
		return Event{}, errors.New("Property 'Service' not found")
	}

	ID := service.(string) + "/" + host.(string)

	status := obj["Status"]
	if status == nil {
		status = obj["State"]
	}
	if status == nil {
		return Event{}, errors.New("Property 'Status' not found")
	}

	checkTTLObj := obj["CheckTTL"]
	if checkTTLObj == nil {
		checkTTLObj = defaultCheckTTL
	}

	checkTTL, ok := checkTTLObj.(float64)
	if !ok {
		return Event{}, errors.New("Property 'checkTTL' should be a number")
	}

	return Event{
		TTL:       checkTTL,
		ID:        ID,
		Status:    status.(string),
		Timestamp: time.Now(),
		Value:     obj,
	}, nil
}
