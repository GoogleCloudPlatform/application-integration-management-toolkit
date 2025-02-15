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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ImportflowCmd to export integrations
var ImportflowCmd = &cobra.Command{
	Use:   "import",
	Short: "Import all versions of an Integration flows to a region",
	Long:  "Import all versions of an Integration flows to a region",
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

		const maxConnections = 4
		name := utils.GetStringParam(cmd.Flag("name"))

		if err = apiclient.FolderExists(folder); err != nil {
			return err
		}
		return integrations.ImportFlow(name, folder, maxConnections)
	},
}

func init() {
	var name string

	ImportflowCmd.Flags().StringVarP(&name, "name", "n",
		"", "Name of the Integration flow")

	ImportflowCmd.Flags().StringVarP(&folder, "folder", "f",
		"", "Folder to import Integration flows")

	_ = ImportflowCmd.MarkFlagRequired("folder")
	_ = ImportflowCmd.MarkFlagRequired("name")
}
