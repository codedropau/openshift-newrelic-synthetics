package cleanup

import (
	"github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/entities"
	routev1 "github.com/openshift/api/route/v1"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	entityutils "github.com/universityofadelaide/openshift-newrelic-synthetics/internal/newrelic/entity"
	routeutils "github.com/universityofadelaide/openshift-newrelic-synthetics/internal/openshift/route"
)

type command struct {
	NewRelicAPIKey      string
	KubernetesMasterURL string
	KubernetesConfig    string
	DryRun              bool
	Namespace           string
}

func syncSynthetics(client *newrelic.NewRelic, routes []routev1.Route, dryRun bool) error {
	entities, err := client.Entities.SearchEntities(entities.SearchEntitiesParams{
		Type: entityutils.TypeMonitor,
	})
	if err != nil {
		return err
	}

	for _, entity := range entities {
		logger := log.WithFields(log.Fields{
			"name": entity.Name,
		})

		tags, err := client.Entities.ListTags(entity.GUID)
		if err != nil {
			return err
		}

		namespace, name, err := entityutils.GetNamespaceName(tags)
		if err != nil {
			logger.Error(err)
			continue
		}

		if exists(routes, namespace, name) {
			logger.Infoln("Skipping. Monitor still has a corresponding OpenShift Route.")
			continue
		}

		if dryRun {
			logger.Infoln("Dry run is enabled. A monitor would have been deleted.")
			continue
		}

		err = client.Synthetics.DeleteMonitor(entity.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func exists(routes []routev1.Route, namespace, name string) bool {
	for _, route := range routes {
		if route.ObjectMeta.Namespace != namespace {
			continue
		}

		if route.ObjectMeta.Name != name {
			continue
		}

		return true
	}

	return false
}

func (cmd *command) run(c *kingpin.ParseContext) error {
	routes, err := routeutils.List(cmd.KubernetesMasterURL, cmd.KubernetesConfig, cmd.Namespace)
	if err != nil {
		return err
	}

	client, err := newrelic.New(newrelic.ConfigPersonalAPIKey(cmd.NewRelicAPIKey))
	if err != nil {
		return err
	}

	return syncSynthetics(client, routes, cmd.DryRun)
}

// Command which executes a command for an environment.
func Command(app *kingpin.Application) {
	c := new(command)

	command := app.Command("cleanup", "Cleanup New Relic Synthetics monitors if OpenShift Routes do not exist").Action(c.run)

	command.Flag("new-relic-api-key", "API key for authenticating with New Relic").Envar("NEW_RELIC_API_KEY").Required().StringVar(&c.NewRelicAPIKey)

	command.Flag("kubernetes-master-url", "URL of the Kubernetes master").Envar("KUBERNETES_MASTER_URL").StringVar(&c.KubernetesMasterURL)
	command.Flag("kubernetes-config", "Path to the Kubernetes config file").Envar("KUBERNETES_CONFIG").StringVar(&c.KubernetesConfig)

	command.Flag("dry-run", "Print out information which would have been executed").Envar("DRY_RUN").BoolVar(&c.DryRun)

	command.Arg("namespace", "").Required().StringVar(&c.Namespace)
}
