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

package integrations

import (
	"encoding/json"
	"os"
	"path"

	"github.com/srinandan/integrationcli/apiclient"
	"github.com/srinandan/integrationcli/client/authconfigs"
	"github.com/srinandan/integrationcli/client/connections"
	"github.com/srinandan/integrationcli/client/integrations"
	"github.com/srinandan/integrationcli/cmd/utils"

	"github.com/spf13/cobra"
)

// ScaffoldCmd to publish an integration flow version
var ScaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Create a scaffolding for the integration flow",
	Long:  "Create a scaffolding for the integration flow and dependencies",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		if err = validate(); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		var integrationBody, overridesBody []byte

		apiclient.SetPrintOutput(false)

		if folder == "" {
			folder, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		if version != "" {
			if integrationBody, err = integrations.Get(name, version, false, true, false); err != nil {
				return err
			}
			if overridesBody, err = integrations.Get(name, version, false, true, true); err != nil {
				return err
			}
		} else if userLabel != "" {
			if integrationBody, err = integrations.GetByUserlabel(name, userLabel, true, false); err != nil {
				return err
			}
			if overridesBody, err = integrations.GetByUserlabel(name, userLabel, true, true); err != nil {
				return err
			}
		} else if snapshot != "" {
			if integrationBody, err = integrations.GetBySnapshot(name, snapshot, true, false); err != nil {
				return err
			}
			if overridesBody, err = integrations.GetBySnapshot(name, snapshot, true, true); err != nil {
				return err
			}
		}

		if err != nil {
			return err
		}

		if err = generateFolder("src"); err != nil {
			return err
		}

		integrationBody, err = apiclient.PrettifyJson(integrationBody)
		if err != nil {
			return err
		}

		if err = apiclient.WriteByteArrayToFile(path.Join(folder, "src", name+".json"), false, integrationBody); err != nil {
			return err
		}

		if len(overridesBody) > 0 {
			if err = generateFolder("overrides"); err != nil {
				return err
			}
			overridesBody, err = apiclient.PrettifyJson(overridesBody)
			if err != nil {
				return err
			}
			if err = apiclient.WriteByteArrayToFile(path.Join(folder, "overrides", "overrides.json"), false, overridesBody); err != nil {
				return err
			}
		}

		authConfigUuids, err := integrations.GetAuthConfigs(integrationBody)
		if err != nil {
			return err
		}

		if len(authConfigUuids) > 0 {
			if err = generateFolder("authconfigs"); err != nil {
				return err
			}
			for _, authConfigUuid := range authConfigUuids {
				authConfigResp, err := authconfigs.Get(authConfigUuid, true)
				if err != nil {
					return err
				}
				authConfigName := getAuthConfigName(authConfigResp)
				authConfigResp, err = apiclient.PrettifyJson(authConfigResp)
				if err != nil {
					return err
				}
				if err = apiclient.WriteByteArrayToFile(path.Join(folder, "authconfigs", authConfigName+".json"), false, authConfigResp); err != nil {
					return err
				}
			}
		}

		connectors, err := integrations.GetConnections(integrationBody)
		if err != nil {
			return err
		}

		if len(connectors) > 0 {
			if err = generateFolder("connectors"); err != nil {
				return err
			}
			for _, connector := range connectors {
				connectionResp, err := connections.Get(connector, "", true)
				if err != nil {
					return err
				}
				if err = apiclient.WriteByteArrayToFile(path.Join(folder, "connectors", connector+".json"), false, connectionResp); err != nil {
					return err
				}
			}
		}

		if err = apiclient.WriteByteArrayToFile(path.Join(folder, "cloudbuild.yaml"), false, []byte(utils.GetCloudBuildYaml())); err != nil {
			return err
		}

		return

	},
}

func init() {
	ScaffoldCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ScaffoldCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	ScaffoldCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	ScaffoldCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	ScaffoldCmd.Flags().StringVarP(&folder, "folder", "f",
		"", "Folder to generate the skaffolding")

	_ = ScaffoldCmd.MarkFlagRequired("name")
}

func generateFolder(name string) (err error) {
	if err = os.Mkdir(path.Join(folder, name), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func getAuthConfigName(authConfigResp []byte) string {
	var m map[string]string
	_ = json.Unmarshal(authConfigResp, &m)
	return m["displayName"]
}
