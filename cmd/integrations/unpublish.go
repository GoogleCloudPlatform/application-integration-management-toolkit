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
	"internal/apiclient"

	"internal/client/integrations"

	"github.com/spf13/cobra"
)

// UnPublishVerCmd to publish an integration flow version
var UnPublishVerCmd = &cobra.Command{
	Use:   "unpublish",
	Short: "Unpublish an integration flow version",
	Long:  "Unpublish an integration flow version",
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
		if version != "" {
			_, err = integrations.Unpublish(name, version)
		} else if userLabel != "" {
			_, err = integrations.UnpublishUserLabel(name, userLabel)
		} else if snapshot != "" {
			_, err = integrations.UnpublishSnapshot(name, snapshot)
		}
		return

	},
}

func init() {
	var version string

	UnPublishVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	UnPublishVerCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	UnPublishVerCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	UnPublishVerCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")

	_ = UnPublishVerCmd.MarkFlagRequired("name")
}
