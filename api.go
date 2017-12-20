package main

import (
	"errors"
	"sort"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/thbkrkr/toctoc/types"
)

// HandleEvent valids and stores an event representing a state of a service for a given host
func HandleEvent(c *gin.Context) {
	ns := c.Param("ns")

	var data map[string]interface{}
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(400, gin.H{"message": "Fail to parse event"})
		return
	}

	event, err := parseEvent(data)
	if err != nil {
		logrus.WithError(err).WithField("body", data).Error("Fail to parse event")
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := events[ns]; !ok {
		events[ns] = map[string]types.Event{}
	}
	events[ns][event.ID] = event
}

// parseEvent parses a generic JSON object in a structured event
// Host: Host || Node
// Service: Service
// Status: Status || State
func parseEvent(obj map[string]interface{}) (types.Event, error) {
	host := obj["Host"]
	if host == nil {
		host = obj["Node"]
	}
	if host == nil {
		return types.Event{}, errors.New("Property 'Host' not found")
	}
	service := obj["Service"]
	if service == nil {
		return types.Event{}, errors.New("Property 'Service' not found")
	}

	ID := service.(string) + "/" + host.(string)

	status := obj["Status"]
	if status == nil {
		status = obj["State"]
	}
	if status == nil {
		return types.Event{}, errors.New("Property 'Status' not found")
	}

	checkTTLObj := obj["CheckTTL"]
	if checkTTLObj == nil {
		checkTTLObj = defaultCheckTTL
	}

	checkTTL, ok := checkTTLObj.(float64)
	if !ok {
		return types.Event{}, errors.New("Property 'checkTTL' should be a number")
	}

	return types.Event{
		TTL:       checkTTL,
		ID:        ID,
		Status:    status.(string),
		Timestamp: time.Now(),
		Value:     obj,
	}, nil
}

// Health returns all events with status 500 if at least one event is in error
func Health(c *gin.Context) {
	mutex.RLock()
	defer mutex.RUnlock()

	evs := events[c.Param("ns")]

	eventsArr := []interface{}{}
	status := 200
	for _, event := range evs {
		eventsArr = append(eventsArr, event)
		if event.Status == types.StatusKO {
			status = 500
		}
	}

	c.JSON(status, eventsArr)
}

// Services returns the list of services and a map of event states indexed by host and service
func Services(c *gin.Context) {
	mutex.RLock()
	defer mutex.RUnlock()

	evs := events[c.Param("ns")]

	servicesMap := map[string]int{}
	eventsMap := map[string]map[string]types.Event{}
	for _, event := range evs {
		if event.ID == "" {
			logrus.WithField("event", event).Warn("event is nil")
			continue
		}
		host := event.GetHost()
		service := event.GetService()
		if _, ok := eventsMap[host]; !ok {
			eventsMap[host] = map[string]types.Event{}
		}
		servicesMap[service] = 1
		eventsMap[host][service] = event
	}

	services := make([]string, len(servicesMap))
	i := 0
	for service := range servicesMap {
		services[i] = service
		i++
	}
	sort.Strings(services)

	c.JSON(200, gin.H{
		"services": services,
		"status":   eventsMap,
	})
}

// DeleteService deletes all events related to a given service
func DeleteService(c *gin.Context) {
	ns := c.Param("ns")
	service := c.Param("service")
	evs := events[ns]

	mutex.Lock()
	defer mutex.Unlock()

	for id, event := range evs {
		if event.GetService() == service {
			logrus.Info("deleteService for host: ", event.GetHost(), " service: ", event.GetService())
			delete(events[ns], id)
		}
	}

	c.JSON(200, gin.H{
		"message": "Service " + service + " removed",
	})
}

// DeleteHost Delete all events related to a given host
func DeleteHost(c *gin.Context) {
	ns := c.Param("ns")
	host := c.Param("host")
	evs := events[ns]

	mutex.Lock()
	defer mutex.Unlock()

	for id, event := range evs {
		if event.GetHost() == host {
			logrus.Info("deleteHost for host: ", event.GetHost(), " service: ", event.GetService())
			delete(events[ns], id)
		}
	}

	c.JSON(200, gin.H{
		"message": "Host " + host + " removed",
	})
}
