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

// ArchiveVerCmd to archive an integration flow version
var ArchiveVerCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archives an integration flow version",
	Long:  "Archives an integration flow version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		version := cmd.Flag("ver").Value.String()

		if err = apiclient.SetRegion(cmd.Flag("reg").Value.String()); err != nil {
			return err
		}
		if err = validate(version); err != nil {
			return err
		}
		return apiclient.SetProjectID(cmd.Flag("proj").Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		version := cmd.Flag("ver").Value.String()
		name := cmd.Flag("name").Value.String()
		if version != "" {
			_, err = integrations.Archive(name, version)
		} else if userLabel != "" {
			_, err = integrations.ArchiveUserLabel(name, userLabel)
		} else if snapshot != "" {
			_, err = integrations.ArchiveSnapshot(name, snapshot)
		}
		return err
	},
}

func init() {
	var name, version string

	ArchiveVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ArchiveVerCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	ArchiveVerCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	ArchiveVerCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")

	_ = ArchiveVerCmd.MarkFlagRequired("name")
}
