package main

import (
	log "github.com/Sirupsen/logrus"
	"time"
)

func Watch() {
	tick := time.NewTicker(time.Second * time.Duration(tick))

	for range tick.C {
		for ns := range events {
			for _, event := range events[ns] {
				if time.Since(event.Timestamp) > time.Second*time.Duration(timeout) {
					alert(ns, event)
				}
			}
		}
	}
}

func alert(ns string, event Event) {
	mutex.Lock()
	defer mutex.Unlock()

	event.Status = StatusKO
	events[ns][event.ID] = event

	log.WithField("ns", ns).WithField("ID", event.ID).Errorf("No event since %d seconds", timeout)

	if kafkaTopic != "" {
		sendAlertToKafka(event)
	}
}

func sendAlertToKafka(event Event) {
	bytes, err := event.toBytes()
	if err != nil {
		log.WithField("Event", event).Errorf("Fail to marshal alert event")
		return
	}
	go q.SendOn(kafkaTopic, bytes)
}
