# openshift-newrelic-synthetics

Sync OpenShift Routes to New Relic Synthetics monitors

## Features

* Automatically provision New Relic Synthetics monitors using Routes as a source of truth

## Usage

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
docker run newrelic-sync
```
