package sync

import (
	"net/url"
	"os"

	"github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/entities"
	"github.com/newrelic/newrelic-client-go/pkg/synthetics"
	routev1 "github.com/openshift/api/route/v1"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	entityutils "github.com/universityofadelaide/openshift-newrelic-synthetics/internal/newrelic/entity"
	monitorutils "github.com/universityofadelaide/openshift-newrelic-synthetics/internal/newrelic/monitor"
	routeutils "github.com/universityofadelaide/openshift-newrelic-synthetics/internal/openshift/route"
)

type command struct {
	NewRelicAPIKey      string
	NewRelicLocation    string
	MonitorType         string
	KubernetesMasterURL string
	KubernetesConfig    string
	DryRun              bool
	Namespace           string
}

func syncSynthetics(client *newrelic.NewRelic, routes []routev1.Route, location, monitorType string, dryRun bool) error {
	monitors, err := monitorutils.List(client)
	if err != nil {
		return err
	}

	team := getTeamName()

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

		// Add trailing slash to avoid redirects on non-root paths.
		if uri.Path != "" && uri.Path != "/" {
			uri.Path = uri.Path + "/"
		}

		urlString := uri.String()

		logger := log.WithFields(log.Fields{
			"namespace": route.ObjectMeta.Namespace,
			"name":      route.ObjectMeta.Name,
			"url":       urlString,
		})

		// Typically whitelisting is used for limiting traffic which can view the site.
		// @todo, Consider alternatives to skipping routes with a whitelist.
		if _, ok := route.ObjectMeta.Annotations[routeutils.AnnotationIPWhitelist]; ok {
			logger.Infoln("Skipping this route because the following annotation is set:", routeutils.AnnotationIPWhitelist)
			continue
		}

		var status synthetics.MonitorStatusType = synthetics.MonitorStatus.Enabled

		if _, ok := route.ObjectMeta.Annotations[routeutils.NewRelicStatus]; ok {
			logger.Infoln("Monitor disabled by annotation:", routeutils.NewRelicStatus)
			status = synthetics.MonitorStatus.Disabled
		}

		monitor := synthetics.Monitor{
			Name:      urlString,
			Type:      synthetics.MonitorType(monitorType), // @todo, Make configurable.
			Frequency: 1,                                   // @todo, Make configurable.
			URI:       urlString,
			Locations: []string{
				location,
			},
			Status:       status,
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
				Key:    entityutils.TagOpenShiftRouteNamespace,
				Values: []string{route.ObjectMeta.Namespace},
			},
			{
				Key:    entityutils.TagOpenShiftRouteName,
				Values: []string{route.ObjectMeta.Name},
			},
			{
				Key:    entityutils.TagOpenShiftRouteToKind,
				Values: []string{route.Spec.To.Kind},
			},
			{
				Key:    entityutils.TagOpenShiftRouteToName,
				Values: []string{route.Spec.To.Name},
			},
			{
				Key:    entityutils.TagTeamTagName,
				Values: []string{team},
			},
		}
	}

	entities, err := client.Entities.SearchEntities(entities.SearchEntitiesParams{
		Type: entityutils.TypeMonitor,
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

func getTeamName() string {
	team := os.Getenv("UA_TEAM_NAME")
	if len(team) == 0 {
		return entityutils.TagTeamName
	}
	return team
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

	return syncSynthetics(client, routes, cmd.NewRelicLocation, cmd.MonitorType, cmd.DryRun)
}

// Command which executes a command for an environment.
func Command(app *kingpin.Application) {
	c := new(command)

	command := app.Command("sync", "Sync OpenShift Routes to New Relic Synthetics monitors.").Action(c.run)

	command.Flag("new-relic-api-key", "API key for authenticating with New Relic").Envar("NEW_RELIC_API_KEY").Required().StringVar(&c.NewRelicAPIKey)
	command.Flag("new-relic-location", "Location which monitors will be provisioned").Default("AWS_AP_SOUTHEAST_2").StringVar(&c.NewRelicLocation)

	command.Flag("monitor-type", "Type of New Relic Synthetics monitor to use").Default("SIMPLE").Envar("NEW_RELIC_MONITOR_TYPE").StringVar(&c.MonitorType)

	command.Flag("kubernetes-master-url", "URL of the Kubernetes master").Envar("KUBERNETES_MASTER_URL").StringVar(&c.KubernetesMasterURL)
	command.Flag("kubernetes-config", "Path to the Kubernetes config file").Envar("KUBERNETES_CONFIG").StringVar(&c.KubernetesConfig)

	command.Flag("dry-run", "Print out information which would have been executed").Envar("DRY_RUN").BoolVar(&c.DryRun)

	command.Arg("namespace", "Namespace where Routes will be queried").Required().StringVar(&c.Namespace)
}
