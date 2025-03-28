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
	"internal/apiclient"
	"internal/client/connections"
	"internal/clilog"
	"internal/cmd/utils"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CrtCustomVerCmd to create a new connection
var CrtCustomVerCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new custom connection version",
	Long:  "Create a new customer connection version in a region",
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

		name := utils.GetStringParam(cmd.Flag("name"))
		id := utils.GetStringParam(cmd.Flag("id"))
		connectionFile := utils.GetStringParam(cmd.Flag("file"))

		if _, err = os.Stat(connectionFile); err != nil {
			return fmt.Errorf("unable to open file %w", err)
		}

		content, err := os.ReadFile(connectionFile)
		if err != nil {
			return fmt.Errorf("unable to open file %w", err)
		}
		return apiclient.PrettyPrint(connections.CreateCustomVersion(name, id, content, serviceAccountName, serviceAccountProject))
	},
	Example: `Create a custom connection version: ` + GetExample(2),
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
