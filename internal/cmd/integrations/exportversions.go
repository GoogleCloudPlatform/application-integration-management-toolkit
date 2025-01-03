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
	"internal/clilog"
	"internal/cmd/utils"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := utils.GetStringParam(cmd.Flag("name"))
		allVersions, _ := strconv.ParseBool(cmd.Flag("all-versions").Value.String())
		if err = apiclient.FolderExists(folder); err != nil {
			return err
		}

		apiclient.SetExportToFile(folder)
		apiclient.DisableCmdPrintHttpResponse()
		clilog.Warning.Println("API calls to integration.googleapis.com have a quota of 480 per min. " +
			"Running this tool against large list of entities can exhaust the quota. Throttling to 360 per min.")

		_, err = integrations.ListVersions(name, -1, "", "", "", allVersions, true, false)
		return err
	},
}

var (
	folder         string
	numConnections int
)

func init() {
	var name string
	allVersions := true

	ExportVerCmd.Flags().StringVarP(&folder, "folder", "f",
		"", "Folder to export Integration flows")
	ExportVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ExportVerCmd.Flags().BoolVarP(&allVersions, "all-versions", "l",
		true, "Export all versions of the Integration")

	_ = ExportVerCmd.MarkFlagRequired("folder")
	_ = ExportVerCmd.MarkFlagRequired("name")
}
