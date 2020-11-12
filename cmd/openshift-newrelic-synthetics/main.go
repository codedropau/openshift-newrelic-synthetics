package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/codedropau/openshift-newrelic-synthetics/cmd/openshift-newrelic-synthetics/sync"
)

func main() {
	app := kingpin.New("openshift-newrelic-synthetics", "Bridging the gap between OpenShift and New Relic Synthetics")

	sync.Command(app)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
