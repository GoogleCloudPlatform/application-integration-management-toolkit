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
	"internal/clilog"

	"internal/client/integrations"

	"github.com/spf13/cobra"
)

// ExportVerCmd to export integrations
var ExportVerCmd = &cobra.Command{
	Use:   "export",
	Short: "Export Integrations flow versions to a folder",
	Long:  "Export Integrations flow versions to a folder",
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
		if err = apiclient.FolderExists(folder); err != nil {
			return err
		}

		apiclient.SetExportToFile(folder)
		apiclient.DisableCmdPrintHttpResponse()
		clilog.Warning.Println("API calls to integration.googleapis.com have a quota of 480 per min. " +
			"Running this tool against large list of entities can exhaust the quota. Throttling to 360 per min.")

		_, err = integrations.ListVersions(name, -1, "", "", "", true, true, false)
		return err
	},
}

var (
	folder         string
	allVersions    bool
	numConnections int
)

func init() {
	var name string

	ExportVerCmd.Flags().StringVarP(&folder, "folder", "f",
		"", "Folder to export Integration flows")
	ExportVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")

	_ = ExportVerCmd.MarkFlagRequired("folder")
	_ = ExportVerCmd.MarkFlagRequired("name")
}
