package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/thbkrkr/go-utilz/http"
	"github.com/thbkrkr/qli/client"
)

var (
	name      = "toctoc"
	buildDate = "dev"
	gitCommit = "dev"
	port      int

	kafkaTopic string
	tick       int
	timeout    int

	mutex  sync.RWMutex
	events = map[string]map[string]Event{}

	q *client.Qlient

	authNs = "krkr,qaas,faas"
)

func init() {
	flag.IntVar(&port, "port", 4242, "Port")
	flag.StringVar(&kafkaTopic, "kafka-topic", "", "Alert Kafka Topic")
	flag.IntVar(&tick, "tick", 30, "Tick seconds")
	flag.IntVar(&timeout, "timeout", 30, "Health timeout in seconds")
	flag.Parse()
}

func main() {
	go Watch()

	hostname, _ := os.Hostname()

	var err error
	q, err = client.NewClientFromEnv(fmt.Sprintf("qws-%s", hostname))
	if err != nil {
		log.WithError(err).Fatal("Fail to create qlient")
	}
	if q == nil {
		log.WithError(err).Fatal("Fail to create qlient (*)")
	}

	http.API(name, buildDate, gitCommit, port, router)
}

func router(e *gin.Engine) {
	r := e.Group("/r", authMiddleware())
	r.POST("/:ns/event", HandleEvent)
	r.GET("/:ns/services", Services)
	r.GET("/:ns/health", Health)
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ns := c.Param("ns")

		if !isAuthorized(ns) {
			c.AbortWithError(401, errors.New("Authorization failed"))
		}
	}
}

func isAuthorized(ns string) bool {
	authNss := strings.Split(authNs, ",")
	for _, auNs := range authNss {
		if auNs == ns {
			return true
		}
	}
	return false
}

// Health return all events with status 500 if at least one event is in error
func Health(c *gin.Context) {
	mutex.RLock()
	defer mutex.RUnlock()

	evs := events[c.Param("ns")]

	eventsArr := []interface{}{}
	status := 200
	for _, event := range evs {
		eventsArr = append(eventsArr, event)
		if event.Status == StatusKO {
			status = 500
		}
	}

	c.JSON(status, eventsArr)
}
