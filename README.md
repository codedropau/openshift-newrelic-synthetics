# openshift-newrelic-synthetics

Sync OpenShift Routes to New Relic Synthetics monitors

## Features

* Automatically provision New Relic Synthetics monitors using Routes as a source of truth

## Usage

The following command will demonstrate which monitors will be created or skipped.

```bash
openshift-newrelic-synthetics sync --new-relic-api-key=xxxxxxxxxxxxxxx --dry-run my-namespace
```
