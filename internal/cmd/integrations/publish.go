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
	"errors"
	"internal/apiclient"
	"internal/client/integrations"
	"os"

	"github.com/spf13/cobra"
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

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		if err = validate(version); err != nil {
			return err
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		version := cmd.Flag("ver").Value.String()
		name := cmd.Flag("name").Value.String()
		configVarsFile := cmd.Flag("config-vars").Value.String()
		var contents []byte

		if configVarsFile != "" {
			if _, err := os.Stat(configVarsFile); os.IsNotExist(err) {
				return err
			}

			contents, err = os.ReadFile(configVarsFile)
			if err != nil {
				return err
			}
		}

		if version != "" {
			_, err = integrations.Publish(name, version, contents)
		} else if userLabel != "" {
			_, err = integrations.PublishUserLabel(name, userLabel, contents)
		} else if snapshot != "" {
			_, err = integrations.PublishSnapshot(name, snapshot, contents)
		}
		return err
	},
}

func init() {
	var name, version, configVars string

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

	_ = PublishVerCmd.MarkFlagRequired("name")
}

func validate(version string) (err error) {
	switch {
	case version == "" && userLabel == "" && snapshot == "":
		return errors.New("must pass oneOf version, snapshot or user-label")
	case version != "" && (userLabel != "" || snapshot != ""):
		return errors.New("must pass oneOf version, snapshot or user-label")
	case userLabel != "" && (version != "" || snapshot != ""):
		return errors.New("must pass oneOf version, snapshot or user-label")
	case snapshot != "" && (userLabel != "" || version != ""):
		return errors.New("must pass oneOf version, snapshot or user-label")
	}
	return nil
}
