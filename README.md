# integrationcli

[![Go Report Card](https://goreportcard.com/badge/github.com/srinandan/integrationcli)](https://goreportcard.com/report/github.com/srinandan/integrationcli)
[![GitHub release](https://img.shields.io/github/v/release/srinandan/integrationcli)](https://github.com/srinandan/integrationcli/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This is a tool to interact with [Application Integration APIs](https://cloud.google.com/application-integration/docs/reference/rest), [Apigee Integration APIs](https://cloud.google.com/apigee/docs/api-platform/integration/reference/rest) and [Connector APIs](https://cloud.google.com/integration-connectors/docs/reference/rest). The tool lets you manage (Create,Get, List, Update, Delete, Export and Import) Integration entities like integrations, authConfigs etc. 

## Installation

`integrationcli` is a binary and you can download the appropriate one for your platform from [here](https://github.com/srinandan/integrationcli/releases)

NOTE: Supported platforms are:

* Darwin
* Windows
* Linux

Run this script to download & install the latest version (on Linux or Darwin)

```sh
curl -L https://raw.githubusercontent.com/srinandan/integrationcli/master/downloadLatest.sh | sh -
```

## Selecting the endpoint

By default `integrationcli` uses Application Integration endpoints. This can be changed per command through the flag `--apigee-integration=true` or set permanently by leveraging preferences. See the `preferences` section below.

## Available Commands

Here is a [list](./docs/integrationcli.md) of available commands

## Preferences

Users can set a default project and region via preferences and those settings will be used for all subsequent commands

```bash
integrationcli prefs set -p project-name -r region

integrationcli integrations list
```

NOTE: the second command uses the org name from perferences

## Access Token Generation

`integrationcli` can use the service account directly and obtain an access token.

```bash
integrationcli token gen -a serviceaccount.json 
```

Parameters
The following parameters are supported. See Common Reference for a list of additional parameters.

* `--account -a` (required) Service Account in json format

Use this access token for all subsequent calls (token expires in 1 hour)

## Access Token Caching

`integrationcli` caches the OAuth Access token for subsequent calls (until the token expires). The access token is stored in `$HOME/.integrationcli`. This path must be readable/writeable by the `integrationcli` process.

```bash
integrationcli token cache -a serviceaccount.json
```

___

## Support

This is not an officially supported Google product