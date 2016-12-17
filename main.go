package main

import (
	"errors"
	"flag"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/thbkrkr/go-utilz/http"
)

var (
	name      = "toctoc"
	buildDate = "dev"
	gitCommit = "dev"

	port    int
	tick    int
	timeout int
	mutex   sync.RWMutex
	events  = map[string]Event{}
)

func init() {
	flag.IntVar(&port, "port", 4242, "Port")
	flag.IntVar(&tick, "tick", 10, "Tick seconds")
	flag.IntVar(&timeout, "timeout", 10, "Timeout in seconds")

	flag.Parse()
}

func main() {
	go watch()

	http.API(name, buildDate, gitCommit, port, router)
}

func router(r *gin.Engine) {
	r.POST("/event", sendEvent)
	r.GET("/health", getHealth)
}

func watch() {
	tick := time.NewTicker(time.Second * time.Duration(tick))
	for range tick.C {
		for _, event := range events {
			if time.Since(event.Timestamp) > time.Second*time.Duration(timeout) {
				alert(event)
			}
		}
	}
}

func alert(event Event) {
	mutex.Lock()
	defer mutex.Unlock()
	event.Status = "ERROR"
	events[event.ID] = event
	log.WithField("ID", event.ID).Errorf("No event since %d seconds", timeout)
	sendEmail(event)
}

func sendEmail(event Event) {
	// TODO
	log.WithField("ID", event.ID).Error("Send alert mail")
}

type Event struct {
	ID        string
	Status    string
	Timestamp time.Time
	Value     interface{}
}

func sendEvent(c *gin.Context) {
	var obj map[string]interface{}
	err := c.BindJSON(&obj)
	if err != nil {
		c.JSON(400, gin.H{"message": "Fail to bind JSON"})
		return
	}

	ID, err := extractID(obj)
	if err != nil {
		c.JSON(400, gin.H{"message": "Fail to extract ID"})
		return
	}

	ID = c.Request.RemoteAddr + "/" + ID
	event := Event{
		ID:        ID,
		Status:    "OK",
		Timestamp: time.Now(),
		Value:     obj,
	}

	mutex.Lock()
	defer mutex.Unlock()
	events[ID] = event
}

func extractID(obj map[string]interface{}) (string, error) {
	host := obj["Host"]
	if host != nil {
		return host.(string), nil
	}

	return "", errors.New("No field found to extract ID")
}

func getHealth(c *gin.Context) {
	mutex.RLock()
	defer mutex.RUnlock()

	c.JSON(200, events)
}
