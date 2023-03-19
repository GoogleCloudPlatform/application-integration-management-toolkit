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
- id: 'Create connections if not present'
  name: us-docker.pkg.dev/appintegration-toolkit/images/integrationcli-builder:latest
  entrypoint: 'bash'
  args:
    - -c
    - |
      gcloud auth print-access-token > /tmp/token

      #setup preferences
      /tmp/integrationcli prefs set integrationcli prefs set --nocheck=true --apigee-integration=false --reg=$LOCATION --proj=$PROJECT_ID
      /tmp/integrationcli token cache -t $(cat /tmp/token)

      #find connection
      for connection in ./connectors/*.json
      do
        /tmp/integrationcli connectors get -n  $(basename -s .json $connection) 2>&1 >/dev/null
        echo $? > /tmp/result
        if [ $(cat /tmp/result) -ne 0 ]; then
          set -e
          #create the connection
          if [ ${_DEFAULT_SA} = "false" ]; then
            echo " --sa ${_SERVICE_ACCOUNT_NAME}" >> /tmp/cmd
          fi

          if [ ${_ENCRYPTED} = "true" ]; then
            echo " -k locations/$LOCATION/keyRings/${_KMS_RING_NAME}/cryptoKeys/${_KMS_KEY_NAME}" >> /tmp/cmd
          fi

          if [ ${_GRANT_PERMISSION} = "true" ]; then
            echo " --g=true" >> /tmp/cmd
          fi

          /tmp/integrationcli connectors create -n $(basename -s .json $connection) -f $connection $(cat /tmp/cmd) > /tmp/response

          fi
          echo "connector response: " $(cat /tmp/response)
        fi
      done

- id: 'Create authconfigs if not present'
  name: us-docker.pkg.dev/appintegration-toolkit/images/integrationcli-builder:latest
  entrypoint: 'bash'
  args:
    - -c
    - |
      gcloud auth print-access-token > /tmp/token

      #setup preferences
      /tmp/integrationcli prefs set integrationcli prefs set --nocheck=true --apigee-integration=false --reg=$LOCATION --proj=$PROJECT_ID
      /tmp/integrationcli token cache -t $(cat /tmp/token)

      #find authconfigs
      for authconfig in ./authconfigs/*.json
      do
        /tmp/integrationcli authconfigs get -n  $(basename -s .json $authconfig) 2>&1 >/dev/null
        echo $? > /tmp/result
        if [ $(cat /tmp/result) -ne 0 ]; then
          set -e
          #create the authconfig
          if [ ${_ENCRYPTED} = "false" ]; then
            /tmp/integrationcli authconfigs create -f $authconfig > /tmp/response
          else
            /tmp/integrationcli authconfigs create -e $authconfig -k locations/$LOCATION/keyRings/${_KMS_RING_NAME}/cryptoKeys/${_KMS_KEY_NAME} > /tmp/response
          fi
          echo "authconfig response: " $(cat /tmp/response)
        fi
      done


- id: 'Create and publish the integration version'
  name: us-docker.pkg.dev/appintegration-toolkit/images/integrationcli-builder:latest
  entrypoint: 'bash'
  args:
    - -c
    - |
      set -e
      gcloud auth print-access-token > /tmp/token

      ls -A src | sed 's/\.json'$// > /tmp/name
      echo "./src/"$(ls -A src) > /tmp/filename

      echo "name: " $(cat /tmp/name)
      echo "filename: " $(cat /tmp/filename)

      #setup preferences
      /tmp/integrationcli prefs set integrationcli prefs set --nocheck=true --apigee-integration=false --reg=$LOCATION --proj=$PROJECT_ID
      /tmp/integrationcli token cache -t $(cat /tmp/token)

      #create the integration version
      /tmp/integrationcli integrations create -n $(cat /tmp/name) -f $(cat /tmp/filename) -u $SHORT_SHA -o ./overrides/overrides.json > /tmp/response
      echo "integration response: " $(cat /tmp/response)
      basename $(cat /tmp/response | jq -r .name) > /tmp/version
      echo "integration version: " $(cat /tmp/version)

      #publish the integration version
      /tmp/integrationcli integrations versions publish -n $(cat /tmp/name) -v $(cat /tmp/version)

#the name of the service account  to use when setting up the connector
substitutions:
  _GRANT_PERMISSIONS: "true"
  _ENCRYPTED: "false"
  _DEFAULT_SA: "false"
  _SERVICE_ACCOUNT_NAME: "connectors"
  _KMS_RING_NAME: "app-integration"
  _KMS_KEY_NAME: "integration"

options:
  logging: CLOUD_LOGGING_ONLY
  substitution_option: "ALLOW_LOOSE"`

func GetCloudBuildYaml() string {
	return cloudBuild
}
