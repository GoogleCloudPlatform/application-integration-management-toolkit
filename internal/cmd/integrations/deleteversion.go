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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// DelVerCmd to delete integration flow version
var DelVerCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an integration flow version",
	Long:  "Delete an integration flow version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		project := utils.GetStringParam(cmd.Flag("proj"))
		region := utils.GetStringParam(cmd.Flag("reg"))
		version := utils.GetStringParam(cmd.Flag("ver"))
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))

		if err = apiclient.SetRegion(region); err != nil {
			return err
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
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		version := utils.GetStringParam(cmd.Flag("ver"))
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))
		name := utils.GetStringParam(cmd.Flag("name"))

		if version != "" {
			_, err = integrations.DeleteVersion(name, version)
		} else if snapshot != "" {
			_, err = integrations.DeleteBySnapshot(name, snapshot)
		} else {
			_, err = integrations.DeleteByUserlabel(name, userLabel)
		}
		return err
	},
}

func init() {
	var name, userLabel, snapshot, version string

	DelVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	DelVerCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	DelVerCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	DelVerCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")

	_ = DelVerCmd.MarkFlagRequired("name")
}
