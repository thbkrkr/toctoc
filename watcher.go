package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"math"
	"net/smtp"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/thbkrkr/toctoc/types"
)

var (
	alertsCache     = map[string]*Alert{}
	alertsCacheLock = sync.RWMutex{}
)

type Alert struct {
	Event     types.Event
	Count     int
	Timestamp time.Time
}

func Watch() {
	tick := time.NewTicker(time.Second * time.Duration(watchTick))
	for range tick.C {
		mutex.Lock()
		for ns := range events {
			for _, event := range events[ns] {
				if event.IsKO() {
					sendAlert(ns, event)
				} else {
					resetAlert(ns, event)
				}
			}
		}
		mutex.Unlock()
	}
}

func resetAlert(ns string, event types.Event) {
	alertsCacheLock.RLock()
	alert := alertsCache[event.ID]
	alertsCacheLock.RUnlock()

	if alert == nil {
		alertsCacheLock.Lock()
		delete(alertsCache, event.ID)
		alertsCacheLock.Unlock()
	}
}

func sendAlert(ns string, event types.Event) {
	event.Status = types.StatusKO
	events[ns][event.ID] = event

	log.WithField("ns", ns).WithField("ID", event.ID).Info("Alert")

	if kafkaAlerter {
		go sendAlertToKafka(event)
	}
	if mailAlerter {
		go sendAlertByMail(event)
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

func sendAlertByMail(event types.Event) error {
	alertsCacheLock.RLock()
	alert := alertsCache[event.ID]
	alertsCacheLock.RUnlock()

	if alert == nil {
		alert = &Alert{
			Event:     event,
			Count:     1,
			Timestamp: time.Now(),
		}
		alertsCacheLock.Lock()
		alertsCache[event.ID] = alert
		alertsCacheLock.Unlock()
	}

	// 1 8 27 64 125 216
	beforeNextAlert := time.Duration(math.Pow(float64(alert.Count), 3)) * time.Minute
	lastAlert := time.Since(alert.Timestamp)

	if lastAlert < beforeNextAlert {
		log.Debugf("Wait next tick to send alert by mail (since=%s < duration=%s)", lastAlert, beforeNextAlert)
		return nil
	}

	log.WithField("ID", event.ID).Debugf("Send alert by mail (since=%s < duration=%s)", lastAlert, beforeNextAlert)

	// Update alert
	alert.Count++
	alert.Timestamp = time.Now()

	alertsCacheLock.Lock()
	alertsCache[event.ID] = alert
	alertsCacheLock.Unlock()

	subject := fmt.Sprintf("%s on host: %s service: %s", event.Status, event.GetHost(), event.GetService())
	message := fmt.Sprintf("%s since %s\n\n%s", subject, time.Since(event.Timestamp), event.GetMessage())

	from := os.Getenv("SMTP_FROM")
	if from == "" {
		return errors.New("SMTP_FROM not defined")
	}
	pass := os.Getenv("SMTP_PASSWORD")
	if from == "" {
		return errors.New("SMTP_PASSWORD not defined")
	}
	host := os.Getenv("SMTP_HOST")
	if from == "" {
		return errors.New("SMTP_HOST not defined")
	}
	to := os.Getenv("ALERT_EMAIL")
	if to == "" {
		return errors.New("ALERT_EMAIL not defined")
	}

	auth := smtp.PlainAuth(subject, from, pass, host)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		message + "\r\n")

	port := "465" // SSL
	servername := host + ":" + port

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(from); err != nil {
		return err
	}

	if err = c.Rcpt(to); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()

	return nil
}
