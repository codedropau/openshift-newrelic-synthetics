package monitor

import (
	"github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/synthetics"
)

func List(client *newrelic.NewRelic) ([]*synthetics.Monitor, error) {
	var monitors []*synthetics.Monitor

	params := &synthetics.ListMonitorsParams{
		Limit: 50,
	}

	for {
		list, err := client.Synthetics.ListMonitors(params)
		if err != nil {
			return nil, err
		}

		monitors = append(monitors, list.Monitors...)

		if list.Count != params.Limit {
			break
		}

		params.Offset = params.Offset + list.Count
	}

	return monitors, nil
}

func CreateOrUpdate(client *newrelic.NewRelic, monitors []*synthetics.Monitor, monitor synthetics.Monitor) (*synthetics.Monitor, error) {
	if id, exists := Exists(monitors, monitor.Name); exists {
		monitor.ID = id
		return client.Synthetics.UpdateMonitor(monitor)
	}

	return client.Synthetics.CreateMonitor(monitor)
}

// Helper function to check if the monitor already exists.
func Exists(monitors []*synthetics.Monitor, name string) (string, bool) {
	for _, monitor := range monitors {
		if monitor.Name == name {
			return monitor.ID, true
		}
	}

	return "", false
}
