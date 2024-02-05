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
	"errors"
	"fmt"
	"os"
	"path"

	"internal/apiclient"
	"internal/clilog"

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
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")
		version := cmd.Flag("ver").Value.String()

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		if userLabel == "" && version == "" && snapshot == "" {
			return errors.New("at least one of userLabel, version or snapshot must be passed")
		}
		if err = validate(version); err != nil {
			return err
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		const jsonExt = ".json"
		var integrationBody, overridesBody []byte
		version := cmd.Flag("ver").Value.String()
		name := cmd.Flag("name").Value.String()

		apiclient.DisableCmdPrintHttpResponse()

		if folder != "" {
			if stat, err := os.Stat(folder); err != nil || !stat.IsDir() {
				return fmt.Errorf("problem with supplied path, %w", err)
			}
		} else {
			if folder, err = os.Getwd(); err != nil {
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

		clilog.Info.Printf("Storing the Integration: %s\n", name)
		if err = generateFolder("src"); err != nil {
			return err
		}

		integrationBody, err = apiclient.PrettifyJson(integrationBody)
		if err != nil {
			return err
		}

		if err = apiclient.WriteByteArrayToFile(
			path.Join(folder, "src", name+jsonExt),
			false,
			integrationBody); err != nil {
			return err
		}

		if len(overridesBody) > 0 && string(overridesBody) != "{}" {
			clilog.Info.Printf("Found overrides in the integration, storing the overrides file\n")
			if err = generateFolder("overrides"); err != nil {
				return err
			}
			overridesBody, err = apiclient.PrettifyJson(overridesBody)
			if err != nil {
				return err
			}
			if err = apiclient.WriteByteArrayToFile(
				path.Join(folder, "overrides", "overrides.json"),
				false,
				overridesBody); err != nil {
				return err
			}
		}

		authConfigUuids, err := integrations.GetAuthConfigs(integrationBody)
		if err != nil {
			return err
		}

		if len(authConfigUuids) > 0 {
			clilog.Info.Printf("Found authconfigs in the integration\n")
			if err = generateFolder("authconfigs"); err != nil {
				return err
			}
			for _, authConfigUUIDs := range authConfigUuids {
				authConfigResp, err := authconfigs.Get(authConfigUUIDs, true)
				if err != nil {
					return err
				}
				authConfigName := getName(authConfigResp)
				clilog.Info.Printf("Storing authconfig %s\n", authConfigName)
				authConfigResp, err = apiclient.PrettifyJson(authConfigResp)
				if err != nil {
					return err
				}
				if err = apiclient.WriteByteArrayToFile(
					path.Join(folder, "authconfigs", authConfigName+jsonExt),
					false,
					authConfigResp); err != nil {
					return err
				}
			}
		}

		connectors, err := integrations.GetConnectionsWithRegion(integrationBody)
		if err != nil {
			return err
		}

		if len(connectors) > 0 {
			clilog.Info.Printf("Found connectors in the integration\n")
			if err = generateFolder("connectors"); err != nil {
				return err
			}
			for _, connector := range connectors {
				connectionResp, err := connections.GetConnectionDetailWithRegion(connector.Name, connector.Region, "", true, true)
				if err != nil {
					return err
				}
				clilog.Info.Printf("Storing connector %s\n", connector)
				connectionResp, err = apiclient.PrettifyJson(connectionResp)
				if err != nil {
					return err
				}
				if err = apiclient.WriteByteArrayToFile(
					path.Join(folder, "connectors", connector.Name+jsonExt),
					false,
					connectionResp); err != nil {
					return err
				}
			}
		}

		instances, err := integrations.GetSfdcInstances(integrationBody)
		if err != nil {
			return err
		}

		if len(instances) > 0 {
			clilog.Info.Printf("Found sfdc instances in the integration\n")
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
					clilog.Info.Printf("Storing sfdcinstance %s\n", instanceName)
					if err = apiclient.WriteByteArrayToFile(
						path.Join(folder, "sfdcinstances", instanceName+jsonExt),
						false,
						instanceBytes); err != nil {
						return err
					}
					clilog.Info.Printf("Storing sfdcchannel %s\n", channelName)
					if err = apiclient.WriteByteArrayToFile(
						path.Join(folder, "sfdcchannels", instanceName+"_"+channelName+jsonExt),
						false,
						channelBytes); err != nil {
						return err
					}
				}
			}
		}

		if cloudBuild {
			clilog.Info.Printf("Storing cloudbuild.yaml\n")
			if err = apiclient.WriteByteArrayToFile(
				path.Join(folder, "cloudbuild.yaml"),
				false,
				[]byte(utils.GetCloudBuildYaml())); err != nil {
				return err
			}
		}

		return err
	},
}

var (
	cloudBuild bool
	env        string
)

func init() {
	var name, version string

	ScaffoldCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ScaffoldCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	ScaffoldCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	ScaffoldCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	ScaffoldCmd.Flags().BoolVarP(&cloudBuild, "cloud-build", "",
		true, "don't generate cloud build file; default is true")
	ScaffoldCmd.Flags().StringVarP(&folder, "folder", "f",
		"", "Folder to generate the scaffolding")
	ScaffoldCmd.Flags().StringVarP(&env, "env", "e",
		"", "Environment name for te scaffolding")

	_ = ScaffoldCmd.MarkFlagRequired("name")
}

func generateFolder(name string) (err error) {
	if name != "src" {
		if env != "" {
			folder = path.Join(folder, env)
		}
		err = os.Mkdir(path.Join(folder, name), os.ModePerm)
	}
	return err
}

func getName(authConfigResp []byte) string {
	var m map[string]string
	_ = json.Unmarshal(authConfigResp, &m)
	return m["displayName"]
}
