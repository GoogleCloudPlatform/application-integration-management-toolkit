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
	"strconv"

	"internal/apiclient"

	"internal/client/integrations"

	"github.com/spf13/cobra"
)

// GetVerCmd to get integration flow
var GetVerCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an integration flow version",
	Long:  "Get an integration flow version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")
		version := cmd.Flag("ver").Value.String()

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}

		minimal, _ := strconv.ParseBool(cmd.Flag("minimal").Value.String())
		overrides, _ := strconv.ParseBool(cmd.Flag("overrides").Value.String())
		basic, _ := strconv.ParseBool(cmd.Flag("basic").Value.String())
		configVar, _ := strconv.ParseBool(cmd.Flag("config-vars").Value.String())

		if configVar && (overrides || minimal || basic) {
			return errors.New("config-vars cannot be combined with overrides, minimal or basic")
		}

		if snapshot == "" && userLabel == "" && version == "" {
			return errors.New("at least one of snapshot, userLabel and version must be supplied")
		}
		if snapshot != "" && (userLabel != "" || version != "") {
			return errors.New("snapshot cannot be combined with userLabel or version")
		}
		if userLabel != "" && (snapshot != "" || version != "") {
			return errors.New("userLabel cannot be combined with snapshot or version")
		}
		if version != "" && (snapshot != "" || userLabel != "") {
			return errors.New("version cannot be combined with snapshot or version")
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var integrationBody, respBody []byte
		version := cmd.Flag("ver").Value.String()
		name := cmd.Flag("name").Value.String()
		minimal, _ := strconv.ParseBool(cmd.Flag("minimal").Value.String())
		overrides, _ := strconv.ParseBool(cmd.Flag("overrides").Value.String())
		basic, _ := strconv.ParseBool(cmd.Flag("basic").Value.String())
		configVar, _ := strconv.ParseBool(cmd.Flag("config-vars").Value.String())

		if configVar {
			apiclient.DisableCmdPrintHttpResponse()
		}

		if version != "" {
			integrationBody, err = integrations.Get(name, version, basic, minimal, overrides)
		} else if snapshot != "" {
			integrationBody, err = integrations.GetBySnapshot(name, snapshot, basic, minimal, overrides)
		} else {
			integrationBody, err = integrations.GetByUserlabel(name, userLabel, basic, minimal, overrides)
		}
		if err != nil {
			return err
		}
		if configVar {
			apiclient.EnableCmdPrintHttpResponse()
			apiclient.ClientPrintHttpResponse.Set(true)
			respBody, err = integrations.GetConfigVariables(integrationBody)
			if err != nil {
				return err
			}
			if respBody != nil {
				apiclient.PrettyPrint(respBody)
			}
			return nil
		}
		return err
	},
}

func init() {
	var name, version string
	minimal, overrides, basic, configVar := false, false, false, false

	GetVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	GetVerCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	GetVerCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	GetVerCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	GetVerCmd.Flags().BoolVarP(&basic, "basic", "b",
		false, "Returns snapshot and version only")
	GetVerCmd.Flags().BoolVarP(&overrides, "overrides", "o",
		false, "Returns overrides only for integration")
	GetVerCmd.Flags().BoolVarP(&minimal, "minimal", "",
		false, "fields of the Integration to be returned; default is false")
	GetVerCmd.Flags().BoolVarP(&configVar, "config-vars", "",
		false, "Returns config variables for the integration")
	_ = GetVerCmd.MarkFlagRequired("name")
}
