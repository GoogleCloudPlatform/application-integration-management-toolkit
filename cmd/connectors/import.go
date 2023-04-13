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

package connectors

import (
	"internal/apiclient"

	"internal/client/connections"

	"github.com/spf13/cobra"
)

// ImportCmd to export connections
var ImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import connections to a region from a folder",
	Long:  "Import connections to a region from a folder",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.FolderExists(folder); err != nil {
			return err
		}

		return connections.Import(folder, createSecret, wait)
	},
}

func init() {
	ImportCmd.Flags().StringVarP(&folder, "folder", "f",
		"", "Folder to import connections")
	ImportCmd.Flags().BoolVarP(&createSecret, "create-secret", "",
		true, "Create Secret Manager secrets when creating the connection")
	ImportCmd.Flags().BoolVarP(&wait, "wait", "",
		false, "Waits for the connector to finish, with success or error")

	_ = ImportCmd.MarkFlagRequired("folder")
}
