// Copyright 2022 Google LLC
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

// CleanCmd to delete integration versions
var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Deletes undeployed/unused versions of an Integration",
	Long:  "Deletes undeployed/unused versions of an Integration",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := cmd.Flag("name").Value.String()
		return integrations.Clean(name, reportOnly, keepList)
	},
}

var (
	reportOnly bool
	keepList   []string
)

func init() {
	var name string

	CleanCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration name")
	CleanCmd.Flags().BoolVarP(&reportOnly, "report", "",
		true, "Report which integration snapshots will be deleted")
	CleanCmd.Flags().StringArrayVarP(&keepList, "keepList", "k",
		[]string{}, "List of snapshots to keep, -k 1 -k 2")

	_ = CleanCmd.MarkFlagRequired("name")
}
