package purge

import (
	"fmt"
	"strings"

	"github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/synthetics"
	"gopkg.in/alecthomas/kingpin.v2"
)

type command struct {
	NewRelicAPIKey        string
	NewRelicMonitorPrefix string
	DryRun                bool
}

const (
	// AnnotationIPWhitelist used when for skipping routes.
	AnnotationIPWhitelist = "haproxy.router.openshift.io/ip_whitelist"
)

func (cmd *command) run(c *kingpin.ParseContext) error {
	nrClient, err := newrelic.New(newrelic.ConfigPersonalAPIKey(cmd.NewRelicAPIKey))
	if err != nil {
		panic(err)
	}

	var monitors []*synthetics.Monitor

	listMonitorsParams := &synthetics.ListMonitorsParams{
		Limit: 50,
	}

	for {
		list, err := nrClient.Synthetics.ListMonitors(listMonitorsParams)
		if err != nil {
			panic(err)
		}

		monitors = append(monitors, list.Monitors...)

		if list.Count != listMonitorsParams.Limit {
			break
		}

		listMonitorsParams.Offset = listMonitorsParams.Offset + list.Count
	}

	for _, monitor := range monitors {
		if !strings.HasPrefix(monitor.Name, cmd.NewRelicMonitorPrefix) {
			continue
		}

		if cmd.DryRun {
			fmt.Println("Dry run is enabled. Would have deleted:", monitor.Name)
			continue
		}

		fmt.Println("Deleting:", monitor.Name)

		err := nrClient.Synthetics.DeleteMonitor(monitor.ID)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

// Command which executes a command for an environment.
func Command(app *kingpin.Application) {
	c := new(command)

	command := app.Command("purge", "Purge New Relic Synthetics monitors with a prefix").Action(c.run)

	command.Flag("new-relic-api-key", "API key for authenticating with New Relic").Required().StringVar(&c.NewRelicAPIKey)
	command.Flag("new-relic-monitor-prefix", "Prefix applied to all objects which are managed by this application").Required().StringVar(&c.NewRelicMonitorPrefix)

	command.Flag("dry-run", "Print out information which would have been executed").Envar("DRY_RUN").BoolVar(&c.DryRun)
}
