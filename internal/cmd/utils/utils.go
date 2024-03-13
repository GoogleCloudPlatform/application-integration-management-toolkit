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
	"io"
	"os"
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
# gcloud builds submit --config=deploy.yaml --project=project-name --region=us-west1

steps:
- id: 'Apply Integration scaffolding configuration'
  name: us-docker.pkg.dev/appintegration-toolkit/images/integrationcli:latest
  args:
    - integrations
    - apply
    - -f
    - .
    - -u
    - ${SHORT_SHA}
    - --wait=${_WAIT}
    - --reg=${_LOCATION}
    - --proj=${PROJECT_ID}
    - --metadata-token
    - $(cat /tmp/cmd)
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
  - targetId: %s-env
---

apiVersion: deploy.cloud.google.com/v1
kind: Target
metadata:
  name: %s-env
customTarget:
  customTargetType: appint-%s-target
---

apiVersion: deploy.cloud.google.com/v1
kind: CustomTargetType
metadata:
  name: appint-%s-target
customActions:
  renderAction: render-app-integration
  deployAction: deploy-app-integration`

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
- name: render-app-integration
  containers:
  - name: render
    image: gcr.io/google.com/cloudsdktool/google-cloud-cli@sha256:66e2681aa3099b4e517e4cdcdefff8f2aa45d305007124ccdc09686f6712d018
    command: ['/bin/bash']
    args:
      - '-c'
      - |-
        echo "Sample manifest rendered content" > manifest.txt
        gsutil cp manifest.txt $CLOUD_DEPLOY_OUTPUT_GCS_PATH/manifest.txt
        echo {\"resultStatus\": \"SUCCEEDED\", \"manifestFile\": \"$CLOUD_DEPLOY_OUTPUT_GCS_PATH/manifest.txt\"} > results.json
        gsutil cp results.json $CLOUD_DEPLOY_OUTPUT_GCS_PATH/results.json
- name: deploy-app-integration
  containers:
  - name: deploy
    image: us-docker.pkg.dev/appintegration-toolkit/images/integrationcli-deploy:latest
    command: ['sh']
    args:
      - '-c'
      - |-
        integrationcli integrations apply --env=dev --reg=$CLOUD_DEPLOY_LOCATION --proj=$CLOUD_DEPLOY_PROJECT --pipeline=$CLOUD_DEPLOY_DELIVERY_PIPELINE --release=$CLOUD_DEPLOY_RELEASE --target=$CLOUD_DEPLOY_TARGET --metadata-token`

func GetCloudDeployYaml(integrationName string, env string) string {
	if env == "" {
		env = "dev"
	}
	return fmt.Sprintf(cloudDeploy, integrationName, env, env, integrationName, integrationName)
}

func GetSkaffoldYaml() string {
	return skaffold
}

func GetCloudBuildYaml() string {
	return cloudBuild
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
