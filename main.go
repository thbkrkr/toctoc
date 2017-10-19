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
	"github.com/thbkrkr/toctoc/types"
)

var (
	name      = "toctoc"
	buildDate = "dev"
	gitCommit = "dev"
	port      int

	watchTick       int
	defaultCheckTTL float64
	kafkaAlerter    bool

	mutex  sync.RWMutex
	events = map[string]map[string]types.Event{}

	q *client.Qlient

	namespaces string
)

func init() {
	flag.IntVar(&port, "port", 4242, "Port")
	flag.IntVar(&watchTick, "watch-tick", 30, "Tick in seconds to watch check states")
	flag.Float64Var(&defaultCheckTTL, "default-check-ttl", 30, "Check TTL in seconds to consider a check in error")
	flag.BoolVar(&kafkaAlerter, "kafka-alerter", false, "Send alerts to Kafka (required env vars: B, U, P, T)")
	flag.StringVar(&namespaces, "ns", "c1,c2", "Namespaces")
	flag.Parse()
}

func main() {
	go Watch()

	hostname, _ := os.Hostname()

	if kafkaAlerter {
		var err error
		q, err = client.NewClientFromEnv(fmt.Sprintf("qws-%s", hostname))
		if err != nil {
			log.WithError(err).Fatal("Fail to create qlient")
		}
		if q == nil {
			log.WithError(err).Fatal("Fail to create qlient (*)")
		}
	}

	http.API(name, buildDate, gitCommit, port, router)
}

func router(e *gin.Engine) {
	e.GET("/help", func(c *gin.Context) {
		c.JSON(200, []string{
			"POST   /r/:ns/event             HandleEvent",
			"GET    /r/:ns/health            Health",
			"GET    /r/:ns/services          Services",
			"DELETE /r/:ns/service/:service  DeleteService",
			"DELETE /r/:ns/host/:host        DeleteHost",
		})
	})

	r := e.Group("/r", authMiddleware())
	r.POST("/:ns/event", HandleEvent)
	r.GET("/:ns/health", Health)
	r.GET("/:ns/services", Services)
	r.DELETE("/:ns/service/:service", DeleteService)
	r.DELETE("/:ns/host/:host", DeleteHost)
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
	authorizedNss := strings.Split(namespaces, ",")
	for _, authorizedNs := range authorizedNss {
		if authorizedNs == ns {
			return true
		}
	}
	return false
}
