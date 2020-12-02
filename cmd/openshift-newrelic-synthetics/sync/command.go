package sync

import (
	"context"
	"net/url"

	"github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/entities"
	"github.com/newrelic/newrelic-client-go/pkg/synthetics"
	routev1 "github.com/openshift/api/route/v1"
	clientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/codedropau/openshift-newrelic-synthetics/internal/monitorutils"
)

type command struct {
	NewRelicAPIKey      string
	NewRelicLocation    string
	KubernetesMasterURL string
	KubernetesConfig    string
	DryRun              bool
	Namespace           string
}

const (
	// AnnotationIPWhitelist used when for skipping routes.
	AnnotationIPWhitelist = "haproxy.router.openshift.io/ip_whitelist"
)

func getRoutes(master, configPath, namespace string) ([]routev1.Route, error) {
	config, err := clientcmd.BuildConfigFromFlags(master, configPath)
	if err != nil {
		return nil, err
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	routes, err := client.Routes(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return routes.Items, nil
}

func syncSynthetics(client *newrelic.NewRelic, routes []routev1.Route, location string, dryRun bool) error {
	monitors, err := monitorutils.List(client)
	if err != nil {
		return err
	}

	tags := make(map[string][]entities.Tag, len(routes))

	for _, route := range routes {
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
			Name:      urlString,
			Type:      "BROWSER", // @todo, Make configurable.
			Frequency: 1, // @todo, Make configurable.
			URI:       urlString,
			Locations: []string{
				location,
			},
			Status:       "ENABLED", // @todo, Make configurable.
			SLAThreshold: 7, // @todo, Make configurable.
		}

		if dryRun {
			logger.Infoln("Dry run is enabled. A monitor would have been created or updated for this route.")
			continue
		}

		logger.Infoln("Creating/Updating monitor")

		m, err := monitorutils.CreateOrUpdate(client, monitors, monitor)
		if err != nil {
			return err
		}

		tags[m.Name] = []entities.Tag{
			{
				Key: "openshiftRouteNamespace",
				Values: []string{route.ObjectMeta.Namespace},
			},
			{
				Key: "openshiftRouteName",
				Values: []string{route.ObjectMeta.Name},
			},
			{
				Key: "openshiftRouteToKind",
				Values: []string{route.Spec.To.Kind},
			},
			{
				Key: "openshiftRouteToName",
				Values: []string{route.Spec.To.Name},
			},
		}
	}

	entities, err := client.Entities.SearchEntities(entities.SearchEntitiesParams{
		Type: "MONITOR",
	})
	if err != nil {
		return err
	}

	for _, entity := range entities {
		log.Infoln("Applying tags to:", entity.Name)

		if _, ok := tags[entity.Name]; !ok {
			continue
		}

		err = client.Entities.AddTags(entity.GUID, tags[entity.Name])
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmd *command) run(c *kingpin.ParseContext) error {
	routes, err := getRoutes(cmd.KubernetesMasterURL, cmd.KubernetesConfig, cmd.Namespace)
	if err != nil {
		return err
	}

	client, err := newrelic.New(newrelic.ConfigPersonalAPIKey(cmd.NewRelicAPIKey))
	if err != nil {
		return err
	}

	return syncSynthetics(client, routes, cmd.NewRelicLocation, cmd.DryRun)
}

// Command which executes a command for an environment.
func Command(app *kingpin.Application) {
	c := new(command)

	command := app.Command("sync", "Sync OpenShift Routes to New Relic Synthetics monitors.").Action(c.run)

	command.Flag("new-relic-api-key", "API key for authenticating with New Relic").Required().StringVar(&c.NewRelicAPIKey)
	command.Flag("new-relic-location", "Location which monitors will be provisioned").Default("AWS_AP_SOUTHEAST_2").StringVar(&c.NewRelicLocation)

	command.Flag("kubernetes-master-url", "URL of the Kubernetes master").Envar("KUBERNETES_MASTER_URL").StringVar(&c.KubernetesMasterURL)
	command.Flag("kubernetes-config", "Path to the Kubernetes config file").Envar("KUBERNETES_CONFIG").StringVar(&c.KubernetesConfig)

	command.Flag("dry-run", "Print out information which would have been executed").Envar("DRY_RUN").BoolVar(&c.DryRun)

	command.Arg("namespace", "").Required().StringVar(&c.Namespace)
}
