# Automate via Cloud Build

This sample repository demonstrates how one could build a CI/CD pipeline for [Application Integration](https://cloud.google.com/application-integration/docs/overview) with [Cloud Build](https://cloud.google.com/build/docs).

## Instructions

There are a few assumptions made in the included [cloudbuild.yaml](./cloudbuild.yaml) file:

* The integration version is in the `src` folder. A sample integration version JSON is included [here](./src/sample.json)
* The integration name is the file name
* The overrides file is in the `overrides` folder can named `overrides.json`
* If included, the authconfig file is in the `authconfig` folder and named `authconfig.json`
* If included, define the connectors in the `connectors` folder

## Configuring Cloud Build

In the setting page of Cloud Build enable the following service account permissions:
* Secret Manager (Secret Manager Accessor)
* Service Accounts (Service Account User)
* Cloud Build (Cloud Build WorkerPool User)
* Cloud KMS (Cloud KMS CryptoKey Decrypter)

Grant the Application Integration Admin role to the Cloud Build Service Agent

```
    gcloud projects add-iam-policy-binding PROJECT_ID \
        --member="serviceAccount:service-PROJECT_NUMBER@gcp-sa-cloudbuild.iam.gserviceaccount.com" \
        --role="roles/integrations.integrationAdmin"
```

## Steps

1. Download the integration from the UI or using `integrationcli`. Here is an example to download via CLI:

```sh

token=$(gcloud auth print-access-token)
integrationcli integrations versions get -n <integration-name> -v <version> -p <project-id> -r <region-name> -t $token > ./src/<integration-name>.json
```

You can also download via a snapshot number like this:

```sh

token=$(gcloud auth print-access-token)
integrationcli integrations versions get -n <integration-name> -s <snapshot> -p <dev-project-id> -r <region-name> -t $token > ./src/<integration-name>.json
```

2. Author overrides (specific for the environment) and store them in the overrides folder. Here is an example overrides for the URL in the REST task

```json
{
    "task_overrides": [{
        "taskId": "1",
        "task:": "GenericRestV2Task",
        "parameters":  {
            "url": {
                "key": "url",
                "value": {
                    "stringValue": "https://httpbin.org/ip"
                }
            }
        }
    }]
}
```

3. Trigger the build manually

```sh

gcloud builds submit --config=cloudbuild.yaml --region=<region-name> --project=<qa-project-name>
```

The integration is labeled with the `SHORT_SHA`, the first seven characters of the commit id

## Overrides

The overrides file contains configuration that is specific to an environment. The structure of the file is as follows:

```json
{
    #trigger overrides can be skipped if API trigger is used
    "trigger_overrides": [{
        "triggerNumber": "1",
        "triggerType": "CLOUD_PUBSUB_EXTERNAL",
        "projectId": "my-project",
        "topicName": "topic"
    }],
    #add task specific overrides here
    "task_overrides": [{
        "taskId": "1",
        "task": "GenericRestV2Task",
        "parameters":  {
            //add parameters to override here
        }
    }]
    #the connector task is handled separately. Add connector overrides here.
    "connection_overrides": [{
        "taskId": "1",
        "task": "GenericConnectorTask",
        "parameters": {
            //add parameters to override here
        }
    }]
}
```

For each override, `taskId` and `task` mandatory. `task` is the task type. Note the configuration settings for the connector task is separated from the rest of the tasks. You will find more samples [here](../test)

### Auth Config Overrides

Auth Configs must be created in each GCP project. The auth config name (which contains the version) different in each project. To override the auth config so it works in the new project, specify the auth config name in the overrides. Here is an example:

```yaml
{
    "task_overrides": [{
        "taskId": "1",
        "task": "CloudFunctionTask",
        "parameters":  {
            "TriggerUrl": {
                "key": "TriggerUrl",
                "value": {
                    "stringValue": "https://region-project.cloudfunctions.net/helloWorld"
                }
            },
            "authConfig": {
                "key": "authConfig",
                "value": {
                    "stringValue": "auth-config-name"
                }
            }
        }
    }]
}
```

### Encrypted Auth Config

If one wishes to store auth config in the source code repo, the file can be encrypted (and base64 encoded) and stored in the repo. To check in an encrypted auth config file, encrypt the clear text auth config file as follows:

```bash
gcloud kms encrypt --plaintext-file=<path-clear-text-authconfig-json> --keyring <key-ring-name> --project <project-id> --location <location> --ciphertext-file=<encrypted-file-name> --key=<kms-key-name>
```

To use this encrypted file in the automation, add the following lines to cloudbuild.yaml

```bash
/tmp/integrationcli authconfigs create -n <auth-config-name> -e <path-to-encrypted-file> -k <cloud-kms-decryption-key-name>
```

## Customize Cloud Builder

This repo uses a custom cloud builder. The cloud builder is hosted at `ghcr.io/srinandan/integrationcli-builder:latest`. The cloud builder can be customized from

1. The [cloud-builder.yaml](../cloud-builder.yaml) file
2. The [Dockerfile](../Dockerfile.builder)

```sh

git clone https://github.com/srinandan/integrationcli.git
gcloud builds submit --config=cloud-builder.yaml --project=project-name
```

Be sure to modify the [cloud-builder.yaml](../cloud-builder.yaml) file to point to the appropriate repo.

___

## Support

This is not an officially supported Google product
