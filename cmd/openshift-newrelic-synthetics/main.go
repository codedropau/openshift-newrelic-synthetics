package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/universityofadelaide/openshift-newrelic-synthetics/cmd/openshift-newrelic-synthetics/cleanup"
	"github.com/universityofadelaide/openshift-newrelic-synthetics/cmd/openshift-newrelic-synthetics/sync"
)

func main() {
	app := kingpin.New("openshift-newrelic-synthetics", "Bridging the gap between OpenShift and New Relic Synthetics")

	cleanup.Command(app)
	sync.Command(app)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
