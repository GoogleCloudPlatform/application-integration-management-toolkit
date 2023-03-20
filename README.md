# integrationcli

[![Go Report Card](https://goreportcard.com/badge/github.com/GoogleCloudPlatform/application-integration-management-toolkit)](https://goreportcard.com/report/github.com/GoogleCloudPlatform/application-integration-management-toolkit)
[![GitHub release](https://img.shields.io/github/v/release/GoogleCloudPlatform/application-integration-management-toolkit)](https://github.com/GoogleCloudPlatform/application-integration-management-toolkit/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This is a tool to interact with [Application Integration APIs](https://cloud.google.com/application-integration/docs/reference/rest), [Apigee Integration APIs](https://cloud.google.com/apigee/docs/api-platform/integration/reference/rest) and [Connector APIs](https://cloud.google.com/integration-connectors/docs/reference/rest). The tool lets you manage (Create,Get, List, Update, Delete, Export and Import) Integration entities like integrations, authConfigs etc.

## Installation

`integrationcli` is a binary and you can download the appropriate one for your platform from [here](https://github.com/GoogleCloudPlatform/application-integration-management-toolkit/releases)

NOTE: Supported platforms are:

* Darwin
* Windows
* Linux

Run this script to download & install the latest version (on Linux or Darwin)

```sh
curl -L https://raw.githubusercontent.com/GoogleCloudPlatform/application-integration-management-toolkit/main/downloadLatest.sh | sh -
```


## Getting Started

### User Tokens
The simplest way to get started with integrationcli is

```sh
token=$(gcloud auth print-access-token)
project=$(gcloud config get-value project | head -n 1)
region=<set region here>

integrationcli integrations list -p $project -r $region -t $token
```

### Set Preferences
If you are using the same GCP project for Integration, then consider setting up preferences so they don't have to be included in every command

```sh
project=$(gcloud config get-value project | head -n 1)
region=<set region here>

integrationcli prefs set --reg=$region --proj=$project
```

Subsequent commands can be like this:

```sh
token=$(gcloud auth print-access-token)
integrationcli integrations list -t $token
```

### Access Token Generation

`integrationcli` can use the service account directly and obtain an access token.

```bash
integrationcli token gen -a serviceaccount.json
```

Parameters
The following parameters are supported. See Common Reference for a list of additional parameters.

* `--account -a` (required) Service Account in json format

Use this access token for all subsequent calls (token expires in 1 hour)

### Access Token Caching

`integrationcli` caches the OAuth Access token for subsequent calls (until the token expires). The access token is stored in `$HOME/.integrationcli`. This path must be readable/writeable by the `integrationcli` process.

```bash
integrationcli token cache -a serviceaccount.json
```

or

```bash
token=$(gcloud auth print-access-token)
integrationcli token cache -t $token
```

## Available Commands

Here is a [list](./docs/integrationcli.md) of available commands

## Selecting the endpoint

By default `integrationcli` uses Application Integration endpoints. This can be changed per command through the flag `--apigee-integration=true` or set permanently by leveraging preferences. See the `preferences` section below.


## Automate via Cloud Build

Please see [here](./cicd/README.md) for details on how to automate deployments via Cloud Build. The container images for integrationcli are:

* Container image for the CLI
```
docker pull us-docker.pkg.dev/appintegration-toolkit/images/integrationcli:latest
```

* Container image for cloud build

```
docker pull us-docker.pkg.dev/appintegration-toolkit/images/integrationcli-builder:latest
```

## Creating Integration Connectors

`integrationcli` can be used to create [Integration Connectors](https://cloud.google.com/integration-connectors/docs). There are two types of Integration Connectors:

### Connectors for Google Managed Applications

Google managed applications include systems like BigQuery, PubSub, Cloud SQL etc. To create a connection for such systems, author a json file like this:

```json
{
    "description": "This is a sample",
    "connectorDetails": {
        "name": "pubsub", ## type of the connector
        "version": 1 ## version is always 1
    },
    "configVariables": [ ## these values are specific to each connector type. this example is for pubsub
        {
            "key": "project_id",
            "stringValue": "$PROJECT_ID$" ## if the project id is the same as the connection, use the variable. Otherwise set the project id explicitly
        },
        {
            "key": "topic_id",
            "stringValue": "mytopic"
        }
    ]
}
```

NOTE: For `ConfigVariables` that take a `region` as a parameter (ex: CloudSQL), you can also use `$REGION$`

Then execute via `integrationcli` like this:

```sh
integrationcli connectors create -n name-of-the-connector -f ./test/pub_sub_connection.json
```

You can optionally pass the service account to be used from the command line:

```sh
integrationcli connectors create -n name-of-the-connector -f ./test/pub_sub_connection.json -sa <sa-name> -sp <sa-project-id>
```

**NOTES:**

* This command assumes the token is cached, otherwise pass the token via `-t`
* If the service account project is not passed and the service account name is passed, then the connection's project id is used
* If the service account doesn't exist, it will be created
* For Google connectors `integrationcli` adds the IAM permissions for the service account to the resource (if the -g flag is passed)

### Connectors for Third Party Applications

Third party application include connectors like Salesforce, Service Now, etc. To create a connection for such systems, author a json file like this:

```yaml
{
    "description": "SFTP Test for demo",
    "connectorDetails": {
        "name": "sftp", ## type of the connector
        "version": 1 ## version is always 1
    },
    "configVariables": [ ## these values are specific to each connector type. this example is for sftp
        {
            "key": "remote_host",
            "stringValue": "example.net"
        },
        {
            "key": "remote_port",
            "stringValue": "22"
        }
    ],
    "authConfig": {
        "authType": "USER_PASSWORD",
        "userPassword": {
            "username": "demo",
            "passwordDetails": {
                "secretName": "sftp-demo", ## this secret is provisioned if it doesn't already exist
                "reference": "./test/password.txt" ## this file contains the data/contents (encrypted or clear) to put in secret manager
            }
        }
    }
}
```

If the connector depends on secret manager, `integrationcli` can create the Secret Manager secret if it is not already provisioned.

Then execute via `integrationcli` like this:

```sh
integrationcli connectors create -n name-of-the-connector -f ./test/sftp_connection.json
```

NOTE: This command assumes the token is cached, otherwise pass the token via `-t`

### Encrypting the Password

When setting the `passwordDetails`, the contents of the password can be encrypted using Cloud KMS

```json
"passwordDetails": {
    "secretName": "sftp-demo",
    "reference": "./test/password.txt" ## the file containing the password - clear text or encrypted
}
```

The file for the password can be in clear text or encrypted text. If encrypted, then a cloud kms key can be passed for decryption. Before storing the file, the file can be encrypted like this:

```sh
gcloud kms encrypt --plaintext-file=./test/password.txt --keyring $key-ring --project $project --location us-west1 --ciphertext-file=enc_passsword.txt --key=$key
base64 ./test/enc_password.txt > ./test/b64_enc_password.txt # on MacOS, use base64 -i ./test/enc_password.txt > ./test/b64_enc_password.txt
```

### Examples of Creating Connectors

* [Big Query](./test/bq_connection.json)
* [Service Now](./test/servicenow_connection.json)
* [Salesforce](./test/salesforce_connections.json)
* [Salesfoce with JWT](./test/salesforce_jwt_connection.json)
* [Oracle](./test/oracle_connection.json)
* [GCS](./test/gcs_connection.json)
* [CloudSQL - MySQL](./test/cloudsql_mysql_connection.json)

___

## Support

This is not an officially supported Google product
