# TocToc

Who is still alive?

## Server

### Start

```bash
Usage of ./toctoc:
  -port int
        Port (default 4242)
  -ns string
        Namespaces (default "c1,c2")
  -default-check-ttl float
        Check TTL in seconds to consider a check in error (default 30)
  -watch-tick int
        Tick in seconds to watch check states (default 30)
  -kafka-alerter
        Send alerts to Kafka (required env vars: B, U, P, T)
```

### API

```bash
> curl io:4242/help -s | jq -r '.[]'
POST   /r/:ns/event             HandleEvent
GET    /r/:ns/health            Health
GET    /r/:ns/services          Services
DELETE /r/:ns/service/:service  DeleteService
DELETE /r/:ns/host/:host        DeleteHost
```

## Ping

Push a state (`OK || KO`) for an host and a given service with a message.

```bash
ping() {
  curl -is "${url}/r/${ns}/event" -XPOST -d '{
    "Host": "n1.b.i.m.io",
    "Service": "badaboum",
    "State": "OK",
    "Message": "Latency < 1ms",
    "CheckTTL": 10
  }'
}
```