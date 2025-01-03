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

// ImportCmd to export integrations
var ImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import all Integration flows from folder",
	Long:  "Import all Integration flows from folder",
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
		const maxConnections = 4
		if err = apiclient.FolderExists(folder); err != nil {
			return err
		}
		apiclient.DisableCmdPrintHttpResponse()
		return integrations.Import(folder, maxConnections)
	},
}

func init() {
	ImportCmd.Flags().StringVarP(&folder, "folder", "f",
		"", "Folder to import Integration flows")

	_ = ImportCmd.MarkFlagRequired("folder")
}
