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

package sfdcchannels

import (
	"internal/apiclient"
	"internal/client/sfdc"
	"internal/clilog"
	"internal/cmd/utils"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GetCmd to get integration flow
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an sfdcchannel in Application Integration",
	Long:  "Get an sfdcchannel in Application Integration",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(utils.GetStringParam(cmdRegion)); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(utils.GetStringParam(cmdProject))
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		instance := utils.GetStringParam(cmd.Flag("instance"))
		name := utils.GetStringParam(cmd.Flag("name"))
		id := utils.GetStringParam(cmd.Flag("id"))
		minimal, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("minimal")))

		if name != "" {
			apiclient.DisableCmdPrintHttpResponse()
			_, respBody, err := sfdc.FindChannel(name, instance)
			if err != nil {
				return err
			}
			apiclient.EnableCmdPrintHttpResponse()
			apiclient.PrettyPrint(respBody)
		} else {
			_, err = sfdc.GetChannel(id, instance, minimal)
		}
		return err
	},
}

func init() {
	var instance, name, id string
	minimal := false

	GetCmd.Flags().StringVarP(&name, "name", "n",
		"", "sfdc channel name")
	GetCmd.Flags().StringVarP(&id, "id", "i",
		"", "sfdc channel uuid")
	GetCmd.Flags().StringVarP(&instance, "instance", "",
		"", "sfdc instance uuid")
	GetCmd.Flags().BoolVarP(&minimal, "minimal", "",
		false, "Minimal number of fields returned; default is false")

	_ = GetCmd.MarkFlagRequired("instance")
}
