package main

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
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

func (e Event) getHost() string {
	return e.Value["Host"].(string)
}

func (e Event) getService() string {
	return e.Value["Service"].(string)
}

func (e Event) toBytes() ([]byte, error) {
	bytes, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func Services(c *gin.Context) {
	mutex.RLock()
	defer mutex.RUnlock()

	evs := events[c.Param("ns")]

	servicesMap := map[string]int{}
	eventsMap := map[string]map[string]interface{}{}
	for _, event := range evs {
		host := event.getHost()
		service := event.getService()
		if _, ok := eventsMap[host]; !ok {
			eventsMap[host] = map[string]interface{}{}
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

func HandleEvent(c *gin.Context) {
	ns := c.Param("ns")

	var data map[string]interface{}
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(400, gin.H{"message": "Fail to parse event"})
		return
	}

	ID, err := extractID(data)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	event := Event{
		ID:        ID,
		Status:    StatusOK,
		Timestamp: time.Now(),
		Value:     data,
	}

	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := events[ns]; !ok {
		events[ns] = map[string]Event{}
	}
	events[ns][ID] = event
}

func extractID(obj map[string]interface{}) (string, error) {
	host := obj["Host"]
	if host == nil {
		return "", errors.New("Property 'Host' not found")
	}
	service := obj["Service"]
	if service == nil {
		return "", errors.New("Property 'Host' not found")
	}

	return service.(string) + "/" + host.(string), nil
}

func DeleteService(c *gin.Context) {
	ns := c.Param("ns")
	service := c.Param("service")
	evs := events[ns]

	mutex.Lock()
	defer mutex.Unlock()

	for id, event := range evs {
		if event.getService() == service {
			logrus.Info("deleteService for host: ", event.getHost(), " service: ", event.getService())
			delete(events[ns], id)
		}
	}

	c.JSON(200, gin.H{
		"message": "Service " + service + " removed",
	})
}

func DeleteHost(c *gin.Context) {
	ns := c.Param("ns")
	host := c.Param("host")
	evs := events[ns]

	mutex.Lock()
	defer mutex.Unlock()

	for id, event := range evs {
		logrus.Info("deleteHost?? ", event.getHost(), " ? ", host)
		if event.getHost() == host {
			logrus.Info("deleteHost for host: ", event.getHost(), " service: ", event.getService())
			delete(events[ns], id)
		}
	}

	c.JSON(200, gin.H{
		"message": "Host " + host + " removed",
	})
}
