package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/thbkrkr/toctoc/types"
)

func Watch() {
	tick := time.NewTicker(time.Second * time.Duration(watchTick))
	for range tick.C {
		for ns := range events {
			for _, event := range events[ns] {
				if time.Since(event.Timestamp) > time.Second*time.Duration(healthTimeout) {
					alert(ns, event)
				}
			}
		}
	}
}

func alert(ns string, event types.Event) {
	mutex.Lock()
	defer mutex.Unlock()

	event.Status = types.StatusKO
	events[ns][event.ID] = event

	log.WithField("ns", ns).WithField("ID", event.ID).Errorf("No event since %d seconds", healthTimeout)

	if kafkaAlerter {
		sendAlertToKafka(event)
	}
}

func sendAlertToKafka(event types.Event) {
	bytes, err := event.ToBytes()
	if err != nil {
		log.WithField("Event", event).Errorf("Fail to marshal alert event")
		return
	}
	go q.Send(bytes)
}
