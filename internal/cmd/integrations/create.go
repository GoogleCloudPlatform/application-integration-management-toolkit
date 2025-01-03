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
	"internal/apiclient"
	"internal/client/integrations"
	"internal/clilog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CreateCmd to list Integrations
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an integration flow with a draft version",
	Long:  "Create an integration flow with a draft version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")
		configVarsJson := cmd.Flag("config-vars-json").Value.String()
		configVarsFile := cmd.Flag("config-vars").Value.String()

		if basic && publish {
			return fmt.Errorf("cannot combine basic and publish flags")
		}

		if !publish && (configVarsFile != "" || configVarsJson != "") {
			return fmt.Errorf("cannot use config-vars and config-vars-json flags when publish is false")
		}

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var overridesContent, contents []byte

		name := cmd.Flag("name").Value.String()
		userLabel := cmd.Flag("user-label").Value.String()
		snapshot := cmd.Flag("snapshot").Value.String()
		configVarsJson := cmd.Flag("config-vars-json").Value.String()
		configVarsFile := cmd.Flag("config-vars").Value.String()

		if configVarsJson == "" {
			if configVarsFile != "" {
				if _, err := os.Stat(configVarsFile); os.IsNotExist(err) {
					return err
				}

				contents, err = os.ReadFile(configVarsFile)
				if err != nil {
					return err
				}
			}
		} else {
			contents = []byte(configVarsJson)
		}
		if _, err := os.Stat(integrationFile); os.IsNotExist(err) {
			return err
		}

		content, err := os.ReadFile(integrationFile)
		if err != nil {
			return err
		}

		if _, err := os.Stat(integrationFile); os.IsNotExist(err) {
			return err
		}

		if overridesFile != "" {
			overridesContent, err = os.ReadFile(overridesFile)
			if err != nil {
				return err
			}
		}

		if publish {
			apiclient.DisableCmdPrintHttpResponse()
		}
		respBody, err := integrations.CreateVersion(name, content, overridesContent, snapshot,
			userLabel, grantPermission, basic)
		if err != nil {
			return err
		}

		if publish {
			apiclient.EnableCmdPrintHttpResponse()
			var integrationMap map[string]interface{}
			err = json.Unmarshal(respBody, &integrationMap)
			if err != nil {
				return err
			}
			version := integrationMap["name"].(string)[strings.LastIndex(integrationMap["name"].(string), "/")+1:]
			if version != "" {
				_, err = integrations.Publish(name, version, contents)
			} else {
				return fmt.Errorf("unable to extract version id from integration")
			}
		}
		return err
	},
	Example: `Create a new Inegration Version with a user label: ` + GetExample(0) + `
Create a new Inegration Version with overrides: ` + GetExample(1) + `
Create a new Inegration Version and publish it: ` + GetExample(2) + `,
Create a new Inegration Version and return a basic response: ` + GetExample(13),
}

var (
	integrationFile, overridesFile  string
	grantPermission, publish, basic bool
)

func init() {
	var name, userLabel, snapshot, configVars, configVarsJson string

	CreateCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	CreateCmd.Flags().StringVarP(&integrationFile, "file", "f",
		"", "Integration flow JSON file path")
	CreateCmd.Flags().StringVarP(&overridesFile, "overrides", "o",
		"", "Integration flow overrides file path")
	CreateCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration version snapshot number")
	CreateCmd.Flags().StringVarP(&userLabel, "userlabel", "u",
		"", "Integration version userlabel")
	CreateCmd.Flags().BoolVarP(&grantPermission, "grant-permission", "g",
		false, "Grant the service account permission for integration triggers; default is false")
	CreateCmd.Flags().BoolVarP(&publish, "publish", "",
		false, "Publish the integration after successful creation; default is false")
	CreateCmd.Flags().BoolVarP(&basic, "basic", "",
		false, "Returns version and snapshot only in the response; default is false")
	CreateCmd.Flags().StringVarP(&configVars, "config-vars", "",
		"", "Path to file containing config variables")
	CreateCmd.Flags().StringVarP(&configVarsJson, "config-vars-json", "",
		"", "Json string containing the config variables if both Json string and file is present Json string will only be used.")

	_ = CreateCmd.MarkFlagRequired("name")
	_ = CreateCmd.MarkFlagRequired("file")
}
