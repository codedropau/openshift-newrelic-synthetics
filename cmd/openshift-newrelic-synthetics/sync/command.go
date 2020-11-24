package sync

import (
	"context"
	"fmt"
	"net/url"

	"github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/synthetics"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type command struct {
	NewRelicAPIKey        string
	NewRelicLocation      string
	NewRelicMonitorPrefix string
	KubernetesMasterURL   string
	KubernetesConfig      string
	DryRun                bool
	Namespace             string
}

const (
	// AnnotationIPWhitelist used when for skipping routes.
	AnnotationIPWhitelist = "haproxy.router.openshift.io/ip_whitelist"
)

func (cmd *command) run(c *kingpin.ParseContext) error {
	config, err := clientcmd.BuildConfigFromFlags(cmd.KubernetesMasterURL, cmd.KubernetesConfig)
	if err != nil {
		panic(err)
	}

	routeClient, err := routev1.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	nrClient, err := newrelic.New(newrelic.ConfigPersonalAPIKey(cmd.NewRelicAPIKey))
	if err != nil {
		panic(err)
	}

	routes, err := routeClient.Routes(cmd.Namespace).List(context.Background(), metav1.ListOptions{})
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

	for _, route := range routes.Items {
		uri := url.URL{
			Scheme: "http", // @todo, Find a cost.
			Host:   route.Spec.Host,
			Path:   route.Spec.Path,
		}

		if route.Spec.TLS != nil {
			uri.Scheme = "https" // @todo, Find a cost.
		}

		urlString := uri.String()

		logger := log.WithFields(log.Fields{
			"namespace": route.ObjectMeta.Namespace,
			"name":      route.ObjectMeta.Name,
			"url":       urlString,
		})

		// Typically whitelisting is used for limiting traffic which can view the site.
		// @todo, Consider alternatives to skipping routes with a whitelist.
		if _, ok := route.ObjectMeta.Annotations[AnnotationIPWhitelist]; ok {
			logger.Infoln("Skipping this route because the following annotation is set:", AnnotationIPWhitelist)
			continue
		}

		monitor := synthetics.Monitor{
			Name:      fmt.Sprintf("%s_%s_%s", cmd.NewRelicMonitorPrefix, route.ObjectMeta.Namespace, route.ObjectMeta.Name),
			Type:      "BROWSER", // @todo, Make configurable.
			Frequency: 1, // @todo, Make configurable.
			URI:       urlString,
			Locations: []string{
				cmd.NewRelicLocation,
			},
			Status:       "ENABLED", // @todo, Make configurable.
			SLAThreshold: 7, // @todo, Make configurable.
		}

		if cmd.DryRun {
			logger.Infoln("Dry run is enabled. A monitor would have been created or updated for this route.")
			continue
		}

		if id, exists := monitorExists(monitors, monitor.Name); exists {
			logger.Infoln("Updating monitor")

			monitor.ID = id

			_, err := nrClient.Synthetics.UpdateMonitor(monitor)
			if err != nil {
				panic(err)
			}

			continue
		}

		logger.Infoln("Creating monitor")

		_, err := nrClient.Synthetics.CreateMonitor(monitor)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

// Helper function to check if the monitor already exists.
func monitorExists(monitors []*synthetics.Monitor, name string) (string, bool) {
	for _, monitor := range monitors {
		if monitor.Name == name {
			return monitor.ID, true
		}
	}

	return "", false
}

// Command which executes a command for an environment.
func Command(app *kingpin.Application) {
	c := new(command)

	command := app.Command("sync", "Sync OpenShift Routes to New Relic Synthetics monitors.").Action(c.run)

	command.Flag("new-relic-api-key", "API key for authenticating with New Relic").Required().StringVar(&c.NewRelicAPIKey)
	command.Flag("new-relic-location", "Location which monitors will be provisioned").Default("AWS_AP_SOUTHEAST_2").StringVar(&c.NewRelicLocation)
	command.Flag("new-relic-monitor-prefix", "Prefix applied to all objects which are managed by this application").Required().StringVar(&c.NewRelicMonitorPrefix)

	command.Flag("kubernetes-master-url", "URL of the Kubernetes master").Envar("KUBERNETES_MASTER_URL").StringVar(&c.KubernetesMasterURL)
	command.Flag("kubernetes-config", "Path to the Kubernetes config file").Envar("KUBERNETES_CONFIG").StringVar(&c.KubernetesConfig)

	command.Flag("dry-run", "Print out information which would have been executed").Envar("DRY_RUN").BoolVar(&c.DryRun)

	command.Arg("namespace", "").Required().StringVar(&c.Namespace)
}
