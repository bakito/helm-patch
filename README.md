[![GoDoc](https://godoc.org/github.com/bakito/helm-patch?status.svg)](http://godoc.org/github.com/bakito/helm-patch)
[![Build Status](https://travis-ci.com/bakito/helm-patch.svg?branch=master)](https://travis-ci.com/bakito/helm-patch)
[![Go Report Card](https://goreportcard.com/badge/github.com/bakito/helm-patch)](https://goreportcard.com/report/github.com/bakito/helm-patch)
[![GitHub Release](https://img.shields.io/github/release/bakito/helm-patch.svg?style=flat)](https://github.com/bakito/helm-patch/releases)

# Helm Patch Plugin

## Overview

This plugin helps fixing helm 3 charts in some szenarios, where default helm 3 might have difficulties.

## Patch API changes

During upgrades of a k2s cluster the version of resources might change. Since the resources are stored in the chart version on the namespace. The versions might become invalid after such an upgrade, since the k2 API might not resolve the resoucde with the old version any more.

This option allows to migrat api version of a certain installend chart version to allow seamless upgrade to the new API.

```console
helm patch api <chart-name> --from v1 --to v2 --kind ConfigMap --revision 1 --dry-run
```

## Adopt existing resources into a new chart

This command allows to adopt / import existing resources into a new chart.
One of the key benefits is, that existing deployments can be seamlessly re-used within a new chart.

```console
helm patch adopt <release-name> <chart> --kind resource-kind --name resource-name
```

## Remove a resources from a new chart

This command allows to remove a resource from a chart.

```console
helm patch rm <chart> --kind resource-kind --name resource-name
```

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbakito%2Fhelm-patch.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbakito%2Fhelm-patch?ref=badge_large)
