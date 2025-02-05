// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"
	"internal/apiclient"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	DefaultFileSplitter = "__"
	LegacyFileSplitter  = "_"
)

const cloudBuild = `# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#to manually trigger from gcloud
# gcloud builds submit --config=cloudbuild.yaml --project=project-name --region=us-west1

steps:
- id: 'Apply Integration scaffolding configuration'
  name: us-docker.pkg.dev/appintegration-toolkit/images/integrationcli:%s
  args:
    - integrations
    - apply
    - -f
    - .
    - -u
    - ${SHORT_SHA}
    - --wait=${_WAIT}
    - --reg=${LOCATION}
    - --proj=${PROJECT_ID}
    - --metadata-token
    # uncomment these as necessary
    #- --g=${_GRANT_PERMISSIONS}
    #- --create-secret=${_CREATE_SECRET}
    #- -k=locations/$LOCATION/keyRings/${_KMS_RING_NAME}/cryptoKeys/${_KMS_KEY_NAME}
    #- --sa=${_SERVICE_ACCOUNT_NAME}

#the name of the service account  to use when setting up the connector
substitutions:
  _GRANT_PERMISSIONS: "true"
  _CREATE_SECRET: "false"
  _ENCRYPTED: "false"
  _DEFAULT_SA: "false"
  _SERVICE_ACCOUNT_NAME: "connectors"
  _KMS_RING_NAME: "app-integration"
  _KMS_KEY_NAME: "integration"
  _WAIT: "true"

options:
  logging: CLOUD_LOGGING_ONLY
  substitution_option: "ALLOW_LOOSE"`

var cloudDeploy = `# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

apiVersion: deploy.cloud.google.com/v1
kind: DeliveryPipeline
metadata:
  name: appint-%s-pipeline
serialPipeline:
  stages:
  - targetId: %s
---

apiVersion: deploy.cloud.google.com/v1
kind: Target
metadata:
  name: %s
customTarget:
  customTargetType: appint-%s-target
deployParameters:
  APP_INTEGRATION_PROJECT_ID: "%s"
  APP_INTEGRATION_REGION: "%s"
---

apiVersion: deploy.cloud.google.com/v1
kind: CustomTargetType
metadata:
  name: appint-%s-target
customActions:
  renderAction: render-%s-integration
  deployAction: deploy-%s-integration`

var skaffold = `# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

apiVersion: skaffold/v4beta7
kind: Config
customActions:
- name: render-%s-integration
  containers:
  - name: render
    image: us-docker.pkg.dev/appintegration-toolkit/images/integrationcli:v%s
    command: ['sh']
    args:
      - '-c'
      - |-
        integrationcli render --output-gcs-path=$CLOUD_DEPLOY_OUTPUT_GCS_PATH
- name: deploy-%s-integration
  containers:
  - name: deploy
    image: us-docker.pkg.dev/appintegration-toolkit/images/integrationcli:v%s
    command: ['sh']
    args:
      - '-c'
      - |-
        integrationcli integrations apply --env=$CLOUD_DEPLOY_TARGET --reg=$CLOUD_DEPLOY_LOCATION --proj=$APP_INTEGRATION_PROJECT_ID --reg=$APP_INTEGRATION_REGION --cloud-deploy=true --run-tests=true --wait=true --metadata-token`

var githubActionApply = `# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# this github action publishes a new integration version
# it also includes any overrides present in overrides.json and config-vars files.
# this sample is using the example in samples/scaffold-example

name: apply-%s-action
permissions: read-all

# Controls when the workflow will run
on: push

env:
  ENVIRONMENT: 'dev'
  PROJECT_ID: ${{ vars.PROJECT_ID }}
  REGION: ${{ vars.REGION }}
  WORKLOAD_IDENTITY_PROVIDER_NAME: ${{ vars.PROVIDER_NAME }}
  SERVICE_ACCOUNT: ${{ vars.SERVICE_ACCOUNT }}

jobs:

  integrationcli-action:

    permissions:
      contents: 'read'
      id-token: 'write'

    name: Apply integration version
    runs-on: ubuntu-latest
    timeout-minutes: 20

    steps:
      - name: Checkout Code
        uses: actions/checkout@1e31de5234b9f8995739874a8ce0492dc87873e2 #v4

      - name: Authenticate Google Cloud
        id: 'gcp-auth'
        uses: google-github-actions/auth@6fc4af4b145ae7821d527454aa9bd537d1f2dc5f #v2.1.7
        with:
          workload_identity_provider: '${{ env.WORKLOAD_IDENTITY_PROVIDER_NAME }}'
          service_account: '${{ env.SERVICE_ACCOUNT }}'
          token_format: 'access_token'

      - name: Calculate variables
        id: 'calc-vars'
        run: |
          echo "SHORT_SHA=$(git rev-parse --short $GITHUB_SHA)" >> $GITHUB_OUTPUT

      - name: Create and Publish Integration
        id: 'publish-integration'
        uses: docker://us-docker.pkg.dev/appintegration-toolkit/images/integrationcli:%s #pin to version of choice
        with:
          args: integrations apply --env=${{ env.ENVIRONMENT}} --folder=. --userlabel=${{ steps.calc-vars.outputs.SHORT_SHA }} --wait=true --proj=${{ env.PROJECT_ID }} --reg=${{ env.REGION }} --token ${{ steps.gcp-auth.outputs.access_token }}`

func GetCloudDeployYaml(integrationName string, env string) string {
	if env == "" {
		env = "dev"
	}
	return fmt.Sprintf(cloudDeploy, integrationName, env, env, integrationName, apiclient.GetProjectID(), apiclient.GetRegion(),
		integrationName, integrationName, integrationName)
}

func GetSkaffoldYaml(integrationName string) string {
	v, _, _ := apiclient.GetBuildParams()
	return fmt.Sprintf(skaffold, integrationName, v, integrationName, v)
}

func GetCloudBuildYaml() string {
	v, _, _ := apiclient.GetBuildParams()
	return fmt.Sprintf(cloudBuild, v)
}

func ReadFile(filePath string) (byteValue []byte, err error) {
	userFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer userFile.Close()

	if byteValue, err = io.ReadAll(userFile); err != nil {
		return nil, err
	}
	return byteValue, err
}

func GetStringParam(flag *pflag.Flag) (param string) {
	param = ""
	if flag != nil {
		param = flag.Value.String()
	}
	return param
}

func GetGithubAction(environment string, integrationName string) string {
	var githubAction string
	if environment != "" {
		githubAction = strings.ReplaceAll(githubActionApply, "'dev'", "'"+environment+"'")
	} else {
		githubAction = githubActionApply
	}
	v, _, _ := apiclient.GetBuildParams()
	return fmt.Sprintf(githubAction, integrationName, v)
}

func GetBasicInfo(cmd *cobra.Command, flag string) bool {
	var param, pref bool

	basicInfo := GetStringParam(cmd.Flag(flag))

	if basicInfo != "" {
		param, _ = strconv.ParseBool(basicInfo)
		return param
	}

	pref, _ = strconv.ParseBool(apiclient.GetBasicInfo())

	return pref
}
