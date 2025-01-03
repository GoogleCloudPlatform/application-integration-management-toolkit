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
	"internal/client/integrations"
	"internal/clilog"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// PublishVerCmd to publish an integration flow version
var PublishVerCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish an integration flow version",
	Long:  "Publish an integration flow version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")
		version := cmd.Flag("ver").Value.String()
		userLabel := cmd.Flag("user-label").Value.String()
		snapshot := cmd.Flag("snapshot").Value.String()
		latest, _ := strconv.ParseBool(cmd.Flag("latest").Value.String())

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		if err = validate(version, userLabel, snapshot, latest); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		version := cmd.Flag("ver").Value.String()
		userLabel := cmd.Flag("user-label").Value.String()
		snapshot := cmd.Flag("snapshot").Value.String()
		name := cmd.Flag("name").Value.String()
		configVarsJson := cmd.Flag("config-vars-json").Value.String()
		configVarsFile := cmd.Flag("config-vars").Value.String()

		var contents []byte
		var info string

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

		latest := ignoreLatest(version, userLabel, snapshot)

		if latest {
			apiclient.DisableCmdPrintHttpResponse()
			// list integration versions, order by state=SNAPSHOT, page size = 1 and return basic info
			respBody, err := integrations.ListVersions(name, 1, "", "state=SNAPSHOT",
				"snapshot_number", false, false, true)
			if err != nil {
				return fmt.Errorf("unable to list versions: %v", err)
			}
			version, err = getIntegrationVersion(respBody)
			if err != nil {
				return err
			}
			apiclient.EnableCmdPrintHttpResponse()
			_, err = integrations.Publish(name, version, contents)
			info = "version " + version
		} else if version != "" {
			_, err = integrations.Publish(name, version, contents)
			info = "version " + version
		} else if userLabel != "" {
			_, err = integrations.PublishUserLabel(name, userLabel, contents)
			info = "user label " + userLabel
		} else if snapshot != "" {
			_, err = integrations.PublishSnapshot(name, snapshot, contents)
			info = "snapshot number " + snapshot
		}
		if err == nil {
			clilog.Info.Printf("Integration %s %s published successfully\n", name, info)
		}
		return err
	},
	Example: `Publishes an integration vesion with the highest snapshot in SNAPSHOT state: ` + GetExample(14) + `
Publishes an integration version that matches user supplied snapshot number: ` + GetExample(15),
}

func init() {
	var name, version, userLabel, snapshot, configVars, configVarsJson string
	var latest bool

	PublishVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	PublishVerCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	PublishVerCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	PublishVerCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	PublishVerCmd.Flags().StringVarP(&configVars, "config-vars", "",
		"", "Path to file containing config variables")
	PublishVerCmd.Flags().StringVarP(&configVarsJson, "config-vars-json", "",
		"", "Json string containing the config variables if both Json string and file is present Json string will only be used.")
	PublishVerCmd.Flags().BoolVarP(&latest, "latest", "",
		true, "Publishes the integeration version with the highest snapshot number in SNAPSHOT state; default is true")

	_ = PublishVerCmd.MarkFlagRequired("name")
}

func validate(version string, userLabel string, snapshot string, latest bool) (err error) {
	switch {
	case !latest && (version == "" && userLabel == "" && snapshot == ""):
		return errors.New("must pass oneOf version, snapshot or user-label")
	case !latest && (version != "" && (userLabel != "" || snapshot != "")):
		return errors.New("must pass oneOf version, snapshot or user-label")
	case !latest && (userLabel != "" && (version != "" || snapshot != "")):
		return errors.New("must pass oneOf version, snapshot or user-label")
	case !latest && (snapshot != "" && (userLabel != "" || version != "")):
		return errors.New("must pass oneOf version, snapshot or user-label")
	}
	return nil
}

func ignoreLatest(version string, userLabel string, snapshot string) (latest bool) {
	if version != "" || userLabel != "" || snapshot != "" {
		return false
	}
	return true
}

func getIntegrationVersion(respBody []byte) (string, error) {
	var data map[string]interface{}
	err := json.Unmarshal(respBody, &data)
	if err != nil {
		return "", err
	}
	if data["integrationVersions"] == nil {
		return "", fmt.Errorf("no integration versions were found")
	}
	integrationVersions := data["integrationVersions"].([]interface{})
	firstIntegrationVersion := integrationVersions[0].(map[string]interface{})
	if firstIntegrationVersion["version"].(string) == "" {
		return "", fmt.Errorf("unable to extract version id from integration")
	}
	return firstIntegrationVersion["version"].(string), nil
}
