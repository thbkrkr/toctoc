package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	"github.com/thbkrkr/toctoc/types"
)

const EnvVarPrefix = "TOCTOC"

type Event struct {
	Host    string
	Service string
	State   string
	Message string
}

type TocTocClient struct {
	ServerAddr     string `envconfig:"addr" required:"true"`
	Namespace      string `envconfig:"ns" required:"true"`
	TickInSeconds  int    `envconfig:"period" default:"10"`
	Host           string `envconfig:"host" required:"true"`
	Service        string `envconfig:"service" required:"true"`
	StatusResolver func() (string, error)
}

func NewTocTocClient(status func() (string, error)) (*TocTocClient, error) {
	var client TocTocClient
	err := envconfig.Process(EnvVarPrefix, &client)
	if err != nil {
		return nil, err
	}
	client.StatusResolver = status

	log.WithFields(log.Fields{
		"addr":    client.ServerAddr,
		"ns":      client.Namespace,
		"tick":    client.TickInSeconds,
		"host":    client.Host,
		"service": client.Service,
	}).Info("Start toctoc ping")
	return &client, nil
}

func (c *TocTocClient) Start() {
	tick := time.NewTicker(time.Duration(c.TickInSeconds) * time.Second).C

	c.do()
	for range tick {
		c.do()
	}
}

func (c *TocTocClient) do() {
	err := c.ping()
	if err != nil {
		log.WithError(err).Error("Fail to ping toctoc server")
	}
}

func (c *TocTocClient) ping() error {
	var event = Event{
		Host:    c.Host,
		Service: c.Service,
	}

	message, err := c.StatusResolver()
	if err != nil {
		event.State = types.StatusKO
		event.Message = err.Error()
	} else {
		event.State = types.StatusOK
		event.Message = message
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/r/%s/event", c.ServerAddr, c.Namespace)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Error to post event (status=%d, url=%s)", resp.StatusCode, url)
	}

	return nil
}
