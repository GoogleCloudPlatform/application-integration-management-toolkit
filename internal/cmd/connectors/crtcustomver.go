// Copyright 2024 Google LLC
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
	"fmt"
	"os"

	"internal/apiclient"
	"internal/client/connections"

	"github.com/spf13/cobra"
)

// CrtCustomVerCmd to create a new connection
var CrtCustomVerCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new custom connection version",
	Long:  "Create a new customer connection version in a region",
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
		id := cmd.Flag("id").Value.String()
		connectionFile := cmd.Flag("file").Value.String()

		if _, err = os.Stat(connectionFile); err != nil {
			return fmt.Errorf("unable to open file %w", err)
		}

		content, err := os.ReadFile(connectionFile)
		if err != nil {
			return fmt.Errorf("unable to open file %w", err)
		}
		_, err = connections.CreateCustomVersion(name, id, content, serviceAccountName, serviceAccountProject)
		return err
	},
}

func init() {
	var name, id string

	CrtCustomVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Connection name")
	CrtCustomVerCmd.Flags().StringVarP(&id, "id", "",
		"", "Identifier assigned to the custom connection version")
	CrtCustomVerCmd.Flags().StringVarP(&connectionFile, "file", "f",
		"", "Custom Connection Version details JSON file path")
	CrtCustomVerCmd.Flags().StringVarP(&serviceAccountName, "sa", "",
		"", "Service Account name for the connection; do not include @<project-id>.iam.gserviceaccount.com")
	CrtCustomVerCmd.Flags().StringVarP(&serviceAccountProject, "sp", "",
		"", "Service Account Project for the connection. Default is the connection's project id")

	_ = CrtCustomVerCmd.MarkFlagRequired("name")
}
