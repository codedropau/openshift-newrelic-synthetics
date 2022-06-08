# openshift-newrelic-synthetics

Sync OpenShift Routes to New Relic Synthetics monitors

## Features

* Automatically provision New Relic Synthetics monitors using Routes as a source of truth

## Usage

Visit NewRelic https://one.newrelic.com/launcher/api-keys-ui.launcher

Create a new 'User' key, name doesn't matter.

The following command will demonstrate which monitors will be created or skipped.

```bash
openshift-newrelic-synthetics sync --new-relic-api-key=xxxxxxxxxxxxxxx --dry-run my-namespace
```

Depending on your go environment, you might need to do something like:

```bash
go run github.com/universityofadelaide/openshift-newrelic-synthetics/cmd/openshift-newrelic-synthetics sync etc etc.
```

## Docker
There is a Dockerfile, so its possible to build a docker image as well with:

```bash
docker build -t newrelic-sync .
```

And then can be run with something like:

```bash
docker run -v ~/.kube/config:/root/.kube/config newrelic-sync sync --kubernetes-config=/root/.kube/config --new-relic-api-key=XXXXXXXXXXXXXX --dry-run namespace
```

Use the UA_TEAM_NAME environment variable to control team tag applied to synthetics rules. Default: sapp

## Disable syncing a route

Disable:

```bash
oc annotate route/example one.newrelic.com/synthetics-status=Disabled
```

Enable (delete the annotation):

```bash
oc annotate route/example one.newrelic.com/synthetics-status-
```
