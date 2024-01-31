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
  entrypoint: 'sh'
  args:
    - -c
    - |
      #setup preferences
      touch /tmp/cmd
      integrationcli prefs set integrationcli prefs set --nocheck=true --reg=$LOCATION --proj=$PROJECT_ID
      integrationcli token cache --metadata-token

      if [ ${_DEFAULT_SA} = "false" ]; then
        echo " --sa ${_SERVICE_ACCOUNT_NAME}" >> /tmp/cmd
      fi

      if [ ${_ENCRYPTED} = "true" ]; then
        echo " -k locations/$LOCATION/keyRings/${_KMS_RING_NAME}/cryptoKeys/${_KMS_KEY_NAME}" >> /tmp/cmd
      fi

      if [ ${_GRANT_PERMISSION} = "true" ]; then
        echo " --g=true" >> /tmp/cmd
      fi

      integrationcli integrations apply -f . --wait=${_WAIT} $(cat /tmp/cmd)

#the name of the service account  to use when setting up the connector
substitutions:
  _GRANT_PERMISSIONS: "true"
  _CREATE_SECRET: "false"
  _ENCRYPTED: "false"
  _DEFAULT_SA: "false"
  _SERVICE_ACCOUNT_NAME: "connectors"
  _KMS_RING_NAME: "app-integration"
  _KMS_KEY_NAME: "integration"
  _WAIT: "false"

options:
  logging: CLOUD_LOGGING_ONLY
  substitution_option: "ALLOW_LOOSE"`

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
