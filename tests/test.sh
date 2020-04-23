#!/bin/bash -eu

testpush() {
  n=1 s=a
  curl -is "http://toctoc.c1.banane.ovh/r/faas/event" -XPOST -d '
    {
      "Host": "n'$n'.k.g.i.h.net",
      "Service": "badaboum.'$s'",
      "State": "OK",
      "Message": "Latency < 100ms",
      "CheckTTL": 10
    }'
}

push() {
  curl -is "$url/r/$ns/event" -XPOST -d "$@"
}

ns=krkr
url=${TOCTOC_ADDR:-"localhost:4242"}

main() {
  while true; do

    n=$(shuf -i 1-5 -n 1)
    s=$(shuf -i 1-5 -n 1)
    d=$(shuf -i 1-4 -n 1)

    echo "event n:$n service:$s sleep:$d"

    push '
    {
      "Host": "n'$n'.k.g.i.h.net",
      "Service": "badaboum.'$s'",
      "State": "OK",
      "Message": "Latency < 100ms",
      "CheckTTL": 3
    }'

    sleep 0.$d
  done
}

#main

push '
    {
      "Host": "n1.k.g.i.h.net",
      "Service": "badaboum.2",
      "State": "OK",
      "Message": "Latency < 100ms",
      "CheckTTL": 6
    }'

######
