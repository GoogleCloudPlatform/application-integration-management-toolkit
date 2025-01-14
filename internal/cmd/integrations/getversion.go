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
	"internal/clilog"
	"internal/cmd/utils"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GetVerCmd to get integration flow
var GetVerCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an integration flow version",
	Long:  "Get an integration flow version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := utils.GetStringParam(cmd.Flag("proj"))
		cmdRegion := utils.GetStringParam(cmd.Flag("reg"))
		version := utils.GetStringParam(cmd.Flag("ver"))
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))
		latest, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("latest")))

		if err = apiclient.SetRegion(cmdRegion); err != nil {
			return err
		} else if err = validate(version, userLabel, snapshot, latest); err != nil {
			return err
		}

		minimal, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("minimal")))
		overrides, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("overrides")))
		basic, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("basic")))
		configVar, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("config-vars")))

		if configVar && (overrides || minimal || basic) {
			return errors.New("config-vars cannot be combined with overrides, minimal or basic")
		} else if err = validate(version, userLabel, snapshot, latest); err != nil {
			return err
		}

		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(cmdProject)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		var integrationBody, respBody []byte
		var basic bool

		version := utils.GetStringParam(cmd.Flag("ver"))
		name := utils.GetStringParam(cmd.Flag("name"))
		minimal, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("minimal")))
		overrides, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("overrides")))
		basic = utils.GetBasicInfo(cmd, "basic")
		configVar, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("config-vars")))
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))

		if configVar {
			apiclient.DisableCmdPrintHttpResponse()
		}

		latest := ignoreLatest(version, userLabel, snapshot)
		if latest {
			if version, err = getLatestVersion(name); err != nil {
				return err
			}
		}

		if version != "" {
			integrationBody, err = integrations.Get(name, version, basic, minimal, overrides)
		} else if snapshot != "" {
			integrationBody, err = integrations.GetBySnapshot(name, snapshot, basic, minimal, overrides)
		} else if userLabel != "" {
			integrationBody, err = integrations.GetByUserlabel(name, userLabel, basic, minimal, overrides)
		} else {
			return errors.New("latest version not found. Must pass oneOf version, snapshot or user-label or fix the integration name")
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
	var name, userLabel, snapshot, version, basic string
	minimal, overrides, configVar := false, false, false
	latest := true

	GetVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	GetVerCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	GetVerCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	GetVerCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	GetVerCmd.Flags().StringVarP(&basic, "basic", "b",
		"", "Returns snapshot and version only")
	GetVerCmd.Flags().BoolVarP(&overrides, "overrides", "o",
		false, "Returns overrides only for integration")
	GetVerCmd.Flags().BoolVarP(&minimal, "minimal", "",
		false, "fields of the Integration to be returned; default is false")
	GetVerCmd.Flags().BoolVarP(&configVar, "config-vars", "",
		false, "Returns config variables for the integration")
	GetVerCmd.Flags().BoolVarP(&latest, "latest", "",
		true, "Get the version with the highest snapshot number in SNAPSHOT state. If none found, selects the highest snapshot in DRAFT state; default is true")

	_ = GetVerCmd.MarkFlagRequired("name")
}
