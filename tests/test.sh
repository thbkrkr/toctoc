#!/bin/bash -eu

main() {
	while true; do
		url=localhost:4242
		ns=krkr

		n=$(shuf -i 1-5 -n 1)
		s=$(shuf -i 1-5 -n 1)
		d=$(shuf -i 1-4 -n 1)

		echo "event n:$n service:$s sleep:$d"

		curl -is "$url/r/$ns/event" -XPOST \
				-d '
			{
				"Host":"n'$n'.k.g.i.h.net",
				"Service":"badaboum.'$s'",
				"State":"OK",
				"Message": "Latency < 100ms",
				"Latency": 221
			}'

		sleep $d

	done
}

main