package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/thbkrkr/toctoc/types"
)

var (
	healthTimeoutDuration = time.Second * time.Duration(healthTimeout)
)

func Watch() {
	tick := time.NewTicker(time.Second * time.Duration(watchTick))
	for range tick.C {
		mutex.Lock()
		for ns := range events {
			for _, event := range events[ns] {
				if isKO(event) {
					alert(ns, event)
				}
			}
		}
		mutex.Unlock()
	}
}

func isKO(event types.Event) bool {
	if event.Status == types.StatusKO {
		return true
	}
	if time.Since(event.Timestamp).Seconds() > float64(healthTimeout) {
		return true
	}
	return false
}

func alert(ns string, event types.Event) {
	event.Status = types.StatusKO
	events[ns][event.ID] = event

	log.WithField("ns", ns).WithField("ID", event.ID).Errorf("No event since %d seconds", healthTimeout)

	if kafkaAlerter {
		go sendAlertToKafka(event)
	}
}

func sendAlertToKafka(event types.Event) {
	msg, err := event.ToBytes()
	if err != nil {
		log.WithField("Event", event).Errorf("Fail to marshal alert event")
		return
	}
	_, _, err = q.Send(msg)
	if err != nil {
		log.WithField("Event", event).Errorf("Fail to send event to kafka")
		return
	}
}
