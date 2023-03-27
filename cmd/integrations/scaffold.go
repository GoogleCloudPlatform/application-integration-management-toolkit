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
	"fmt"
	"os"
	"path"

	"internal/apiclient"

	"internal/client/authconfigs"
	"internal/client/connections"
	"internal/client/integrations"
	"internal/client/sfdc"

	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/utils"

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

		if folder != "" {
			if stat, err := os.Stat(folder); err != nil || !stat.IsDir() {
				return fmt.Errorf("problem with supplied path, %v", err)
			}
		} else {
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

		fmt.Printf("Storing the Integration: %s\n", name)
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

		if len(overridesBody) > 0 && string(overridesBody) != "{}" {
			fmt.Printf("Found overrides in the integration, storing the overrides file\n")
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
			fmt.Printf("Found authconfigs in the integration\n")
			if err = generateFolder("authconfigs"); err != nil {
				return err
			}
			for _, authConfigUuid := range authConfigUuids {
				authConfigResp, err := authconfigs.Get(authConfigUuid, true)
				if err != nil {
					return err
				}
				authConfigName := getName(authConfigResp)
				fmt.Printf("Storing authconfig %s\n", authConfigName)
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
			fmt.Printf("Found connectors in the integration\n")
			if err = generateFolder("connectors"); err != nil {
				return err
			}
			for _, connector := range connectors {
				connectionResp, err := connections.Get(connector, "", true, true)
				if err != nil {
					return err
				}
				fmt.Printf("Storing connector %s\n", connector)
				connectionResp, err = apiclient.PrettifyJson(connectionResp)
				if err != nil {
					return err
				}
				if err = apiclient.WriteByteArrayToFile(path.Join(folder, "connectors", connector+".json"), false, connectionResp); err != nil {
					return err
				}
			}
		}

		instances, err := integrations.GetSfdcInstances(integrationBody)
		if err != nil {
			return err
		}

		if len(instances) > 0 {
			fmt.Printf("Found sfdc instances in the integration\n")
			instancesContent, err := sfdc.GetInstancesAndChannels(instances)
			if err != nil {
				return err
			}
			if len(instancesContent) > 0 {
				if err = generateFolder("sfdcinstances"); err != nil {
					return err
				}
				if err = generateFolder("sfdcchannels"); err != nil {
					return err
				}
				for instance, channel := range instancesContent {
					instanceBytes, _ := apiclient.PrettifyJson([]byte(instance))
					channelBytes, _ := apiclient.PrettifyJson([]byte(channel))
					instanceName := getName([]byte(instance))
					channelName := getName([]byte(channel))
					fmt.Printf("Storing sfdcinstance %s\n", instanceName)
					if err = apiclient.WriteByteArrayToFile(path.Join(folder, "sfdcinstances", instanceName+".json"), false, instanceBytes); err != nil {
						return err
					}
					fmt.Printf("Storing sfdcchannel %s\n", channelName)
					if err = apiclient.WriteByteArrayToFile(path.Join(folder, "sfdcchannels", channelName+".json"), false, channelBytes); err != nil {
						return err
					}
				}
			}
		}

		fmt.Printf("Storing cloudbuild.yaml\n")
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

func getName(authConfigResp []byte) string {
	var m map[string]string
	_ = json.Unmarshal(authConfigResp, &m)
	return m["displayName"]
}
