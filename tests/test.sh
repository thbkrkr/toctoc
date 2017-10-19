#!/bin/bash -eu

push() {
  curl -is "$url/r/$ns/event" -XPOST -d "$@"
}

main() {
  while true; do
    url=${TOCTOC_ADDR:-"localhost:4242"}
    ns=krkr

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
      "CheckTTL": 10
    }'

    sleep 0.$d
  done
}

main

######

draft_pipeline() {

  # update the code of a service by pushing new code in a git repo
  # @human -> stash -> kafka

  push '
  {
    "Timestamp": 1498859806,
    "Kind": "gitpush",
    "Node": "io.blurb.space",
    "Service": "badaboum.'$s'",
    "Name": "gitpush a9a32za in badaboum.'$s'",
    "Status": "OK",
    "Output": "Fix bug",
    "Sha1": "a9a32za"
  }'
  # @kafka <- call jenkins -> start build job


  # start a job to build the service
  push '
  {
    "Timestamp": 1498859806,
    "Kind": "build",
    "Node": "j.blurb.space",
    "Service": "badaboum.'$s'",
    "Name": "build a9a32za badaboum.'$s'",
    "Status": "OK",
    "Ouput": "XX...........",
    "Cmd": "docker build -t xxxx .",
    "Sha1": "a9a32za"
  }'
  # @build job result -> kafka
  # @build job result OK -> start deploy job

  # deploy the new version
  push '
  {
    "Timestamp": 1498859806,
    "Kind": "deploy",
    "Node": "n'$n'.blurb.space",
    "Service": "badaboum.'$s'",
    "Name": "deploy a9a32za badaboum.'$s'",
    "Status": "OK",
    "Ouput": "XX...........",
    "Cmd": "docker build -t xxxx .",
    "Sha1": "a9a32za"
  }'
  # @kafka <- call jenkins -> start build job

  # check the heath of the service
  push '
  {
    "Timestamp": 1498859806,
    "Kind": "check",
    "Node": "n'$n'.blurb.space",
    "Service": "badaboum.'$s'",
    "Name": "check-http badaboum.'$s'",
    "Status": "OK",
    "Ouput": "XX...........",
  }'

}
