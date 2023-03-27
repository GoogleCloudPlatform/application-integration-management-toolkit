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

// ExportCmd to export integrations
var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all Integrations flows in a region to a folder",
	Long:  "Export all Integrations flows in a region to a folder",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.FolderExists(folder); err != nil {
			return err
		}

		// check if connections argument was passed, use default value if not
		numConnections, _ := cmd.Flags().GetInt("connections")
		if numConnections > 0 {
			return integrations.ExportConcurrent(folder, numConnections)
		}
		return integrations.Export(folder)

	},
}

func init() {
	ExportCmd.Flags().StringVarP(&folder, "folder", "f",
		"", "Folder to export Integration flows")
	ExportCmd.Flags().IntVarP(&numConnections, "connections", "c",
		-1, "# of concurrent routines to use")

	_ = ExportCmd.MarkFlagRequired("folder")
}
