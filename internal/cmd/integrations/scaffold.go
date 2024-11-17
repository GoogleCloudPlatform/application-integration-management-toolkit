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
	"internal/apiclient"
	"internal/client/authconfigs"
	"internal/client/connections"
	"internal/client/integrations"
	"internal/client/sfdc"
	"internal/clilog"
	"internal/cmd/utils"
	"os"
	"path"
	"path/filepath"

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
		var fileSplitter string
		var integrationBody, overridesBody, testCasesBody []byte
		version := cmd.Flag("ver").Value.String()
		name := cmd.Flag("name").Value.String()

		apiclient.DisableCmdPrintHttpResponse()

		if useUnderscore {
			fileSplitter = utils.LegacyFileSplitter
		} else {
			fileSplitter = utils.DefaultFileSplitter
		}

		if folder != "" {
			if stat, err := os.Stat(folder); err != nil || !stat.IsDir() {
				return fmt.Errorf("problem with supplied path, %w", err)
			}
		} else {
			if folder, err = os.Getwd(); err != nil {
				return err
			}
		}

		baseFolder := folder
		if env != "" {
			folder = path.Join(folder, env)
			if err = generateFolder(folder); err != nil {
				return err
			}
		}

		// Get

		if version != "" {
			if integrationBody, err = integrations.Get(name, version, false, true, false); err != nil {
				return err
			}
			if overridesBody, err = integrations.Get(name, version, false, false, true); err != nil {
				return err
			}
			if testCasesBody, err = integrations.ListTestCases(name, version); err != nil {
				return err
			}
		} else if userLabel != "" {
			if integrationBody, err = integrations.GetByUserlabel(name, userLabel, false, true, false); err != nil {
				return err
			}
			if overridesBody, err = integrations.GetByUserlabel(name, userLabel, false, false, true); err != nil {
				return err
			}
			if testCasesBody, err = integrations.ListTestCasesByUserlabel(name, userLabel); err != nil {
				return err
			}
		} else if snapshot != "" {
			if integrationBody, err = integrations.GetBySnapshot(name, snapshot, false, true, false); err != nil {
				return err
			}
			if overridesBody, err = integrations.GetBySnapshot(name, snapshot, false, false, true); err != nil {
				return err
			}
			if testCasesBody, err = integrations.ListTestCasesBySnapshot(name, snapshot); err != nil {
				return err
			}
		}

		clilog.Info.Printf("Storing the Integration: %s\n", name)
		if err = generateFolder(path.Join(baseFolder, "src")); err != nil {
			return err
		}

		integrationBody, err = apiclient.PrettifyJson(integrationBody)
		if err != nil {
			return err
		}

		if err = apiclient.WriteByteArrayToFile(
			path.Join(baseFolder, "src", name+jsonExt),
			false,
			integrationBody); err != nil {
			return err
		}

		if len(testCasesBody) > 0 {
			clilog.Info.Printf("Found test cases in the integration, storing the test cases file\n")
			if err = generateFolder(path.Join(folder, "testcases")); err != nil {
				return err
			}
			if err = generateTestcases(testCasesBody, folder); err != nil {
				return err
			}
		}

		// write integration overrides
		if len(overridesBody) > 0 && string(overridesBody) != "{}" {
			clilog.Info.Printf("Found overrides in the integration, storing the overrides file\n")
			if err = generateFolder(path.Join(folder, "overrides")); err != nil {
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

		// write integation config variables
		configVariables, err := integrations.GetConfigVariables(integrationBody)
		if err != nil {
			return err
		}
		if len(configVariables) > 0 {
			clilog.Info.Printf("Found config variables in the integration, storing the config file\n")
			if err = generateFolder(path.Join(folder, "config-variables")); err != nil {
				return err
			}
			configVariables, err = apiclient.PrettifyJson(configVariables)
			if err = apiclient.WriteByteArrayToFile(
				path.Join(folder, "config-variables", name+"-config.json"),
				false,
				configVariables); err != nil {
				return err
			}
		}

		// extract code
		if extractCode {
			codeMap, err := integrations.ExtractCode(integrationBody)
			if err != nil {
				return err
			}
			if len(codeMap["JavaScriptTask"]) > 0 {
				javascriptFolder := path.Join(baseFolder, "src", "javascript")
				if err = generateFolder(javascriptFolder); err != nil {
					return err
				}
				clilog.Info.Printf("Found JavaScript files in the integration; generating separate files\n")
				for taskId, taskContent := range codeMap["JavaScriptTask"] {
					if err = apiclient.WriteByteArrayToFile(
						path.Join(javascriptFolder, "javascript_"+string(taskId)+".js"),
						false,
						[]byte(taskContent)); err != nil {
						return err
					}
				}
			}
			if len(codeMap["JsonnetMapperTask"]) > 0 {
				jsonnetFolder := path.Join(baseFolder, "src", "datatransformer")
				if err = generateFolder(jsonnetFolder); err != nil {
					return err
				}
				clilog.Info.Printf("Found Jsonnet files in the integration; generating separate files\n")
				for taskId, taskContent := range codeMap["JsonnetMapperTask"] {
					if err = apiclient.WriteByteArrayToFile(
						path.Join(jsonnetFolder, "datatransformer_"+string(taskId)+".jsonnet"),
						false,
						[]byte(taskContent)); err != nil {
						return err
					}
				}
			}
		}

		// auth config
		authConfigUuids, err := integrations.GetAuthConfigs(integrationBody)
		if err != nil {
			return err
		}

		if !skipAuthconfigs {
			if len(authConfigUuids) > 0 {
				clilog.Info.Printf("Found authconfigs in the integration\n")
				if err = generateFolder(path.Join(folder, "authconfigs")); err != nil {
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
		} else {
			clilog.Info.Printf("Skipping scaffold of authconfigs configuration\n")
		}

		if !skipConnectors {
			connectors, err := integrations.GetConnectionsWithRegion(integrationBody)
			if err != nil {
				return err
			}

			if len(connectors) > 0 {
				clilog.Info.Printf("Found connectors in the integration\n")
				if err = generateFolder(path.Join(folder, "connectors")); err != nil {
					return err
				}
				// check for custom connectors
				for _, connector := range connectors {
					if connector.CustomConnection {
						if err = generateFolder(path.Join(folder, "custom-connectors")); err != nil {
							return err
						}
						break
					}
				}
				for _, connector := range connectors {
					if connector.CustomConnection {
						customConnectionResp, err := connections.GetCustomVersion(connector.Name, connector.Version, true)
						if err != nil {
							return err
						}
						clilog.Info.Printf("Storing custom connector %s\n", connector.Name)
						customConnectionResp, err = apiclient.PrettifyJson(customConnectionResp)
						if err != nil {
							return err
						}
						if err = apiclient.WriteByteArrayToFile(
							path.Join(folder, "custom-connectors", connector.Name+fileSplitter+connector.Version+jsonExt),
							false,
							customConnectionResp); err != nil {
							return err
						}
					} else {
						connectionResp, err := connections.GetConnectionDetailWithRegion(connector.Name, connector.Region, "", true, true)
						if err != nil {
							return err
						}
						clilog.Info.Printf("Storing connector %s\n", connector.Name)
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
			}
		} else {
			clilog.Info.Printf("Skipping scaffold of connector configuration\n")
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
				if err = generateFolder(path.Join(folder, "sfdcinstances")); err != nil {
					return err
				}
				if err = generateFolder(path.Join(folder, "sfdcchannels")); err != nil {
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
						path.Join(folder, "sfdcchannels", instanceName+fileSplitter+channelName+jsonExt),
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
				path.Join(baseFolder, "cloudbuild.yaml"),
				false,
				[]byte(utils.GetCloudBuildYaml())); err != nil {
				return err
			}
		}

		if cloudDeploy {
			clilog.Info.Printf("Storing clouddeploy.yaml and skaffold.yaml\n")
			if err = apiclient.WriteByteArrayToFile(
				path.Join(baseFolder, "clouddeploy.yaml"),
				false,
				[]byte(utils.GetCloudDeployYaml(name, env))); err != nil {
				return err
			}
			if err = apiclient.WriteByteArrayToFile(
				path.Join(baseFolder, "skaffold.yaml"),
				false,
				[]byte(utils.GetSkaffoldYaml())); err != nil {
				return err
			}
		}

		return err
	},
}

var (
	cloudBuild, cloudDeploy, skipConnectors, skipAuthconfigs, useUnderscore, extractCode bool
	env                                                                                  string
)

const jsonExt = ".json"

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
		false, "Generate cloud build file; default is false")
	ScaffoldCmd.Flags().BoolVarP(&cloudDeploy, "cloud-deploy", "",
		false, "Generate cloud deploy files; default is false")
	ScaffoldCmd.Flags().StringVarP(&folder, "folder", "f",
		"", "Folder to generate the scaffolding")
	ScaffoldCmd.Flags().StringVarP(&env, "env", "e",
		"", "Environment name for the scaffolding")
	ScaffoldCmd.Flags().BoolVarP(&skipConnectors, "skip-connectors", "",
		false, "Exclude connectors from scaffold")
	ScaffoldCmd.Flags().BoolVarP(&skipAuthconfigs, "skip-authconfigs", "",
		false, "Exclude authconfigs from scaffold")
	ScaffoldCmd.Flags().BoolVarP(&useUnderscore, "use-underscore", "",
		false, "Use underscore as a file splitter; default is __")
	ScaffoldCmd.Flags().BoolVarP(&extractCode, "extract-code", "x",
		false, "Extract JavaScript and Jsonnet code as separate files")

	_ = ScaffoldCmd.MarkFlagRequired("name")
}

func generateFolder(name string) (err error) {
	if _, err = os.Stat(name); !os.IsNotExist(err) {
		return nil
	}
	err = os.Mkdir(name, os.ModePerm)
	return err
}

func getName(authConfigResp []byte) string {
	var m map[string]string
	_ = json.Unmarshal(authConfigResp, &m)
	return m["displayName"]
}

func generateTestcases(testcases []byte, folder string) error {

	var data map[string]interface{}

	err := json.Unmarshal(testcases, &data)
	if err != nil {
		return fmt.Errorf("Error decoding JSON: %s", err)
	}

	tc := data["testCases"].([]interface{})

	for _, t := range tc {
		obj := t.(map[string]interface{})
		jsonData, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("Error encoding JSON: %s", err)
		}
		name, err := getTestCaseName(obj)
		if err != nil {
			return fmt.Errorf("unable to get name: %v", err)
		}
		jsonData, err = apiclient.PrettifyJson(jsonData)
		if err != nil {
			return err
		}
		if err = apiclient.WriteByteArrayToFile(
			path.Join(folder, "testcases", name+jsonExt),
			false,
			jsonData); err != nil {
			return err
		}
	}
	return nil
}

func getTestCaseName(jsonData map[string]interface{}) (string, error) {
	if name, ok := jsonData["name"].(string); ok && name != "" {
		return filepath.Base(name), nil
	}
	return "", fmt.Errorf("name not found")
}
