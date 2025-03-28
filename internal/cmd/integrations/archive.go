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
	"fmt"
	"internal/apiclient"
	"internal/client/integrations"
	"internal/clilog"
	"internal/cmd/utils"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ArchiveVerCmd to archive an integration flow version
var ArchiveVerCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archives an integration flow version",
	Long:  "Archives an integration flow version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		version := utils.GetStringParam(cmd.Flag("ver"))
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))
		latest, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("latest")))

		if err = apiclient.SetRegion(utils.GetStringParam(cmd.Flag("reg"))); err != nil {
			return err
		}
		if err = validate(version, userLabel, snapshot, latest); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(utils.GetStringParam(cmd.Flag("proj")))
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		version := utils.GetStringParam(cmd.Flag("ver"))
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))
		name := utils.GetStringParam(cmd.Flag("name"))

		var response apiclient.APIResponse

		latest := ignoreLatest(version, userLabel, snapshot)

		if latest {
			// list integration versions, order by state=SNAPSHOT, page size = 1 and return basic info
			response = integrations.ListVersions(name, 1, "", "state=SNAPSHOT",
				"snapshot_number", false, false, true)
			if response.Err != nil {
				return fmt.Errorf("unable to list versions: %v", response.Err)
			}
			if string(response.RespBody) == "{}" {
				if response = integrations.ListVersions(name, 1, "", "state=DRAFT",
					"snapshot_number", false, false, true); response.Err != nil {
					return fmt.Errorf("unable to list versions: %v", response.Err)
				}
			}
			version, err = integrations.GetIntegrationVersion(response.RespBody)
			if err != nil {
				return err
			}
			return apiclient.PrettyPrint(integrations.Archive(name, version))
		} else if version != "" {
			return apiclient.PrettyPrint(integrations.Archive(name, version))
		} else if userLabel != "" {
			return apiclient.PrettyPrint(integrations.ArchiveUserLabel(name, userLabel))
		} else if snapshot != "" {
			return apiclient.PrettyPrint(integrations.ArchiveSnapshot(name, snapshot))
		}
		return err
	},
}

func init() {
	var name, userLabel, snapshot, version string
	var latest bool

	ArchiveVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ArchiveVerCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	ArchiveVerCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	ArchiveVerCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	ArchiveVerCmd.Flags().BoolVarP(&latest, "latest", "",
		true, "Archives the integeration version with the highest snapshot number in SNAPSHOT state; default is true")

	_ = ArchiveVerCmd.MarkFlagRequired("name")
}
