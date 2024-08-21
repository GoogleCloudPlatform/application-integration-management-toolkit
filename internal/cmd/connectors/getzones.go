// Copyright 2023 Google LLC
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
	"strconv"

	"github.com/spf13/cobra"
)

// GetManagedZonesCmd to list Connections
var GetManagedZonesCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a managedzone configuration",
	Long:  "Get a managedzone configuration",
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
		overrides, _ := strconv.ParseBool(cmd.Flag("overrides").Value.String())

		_, err = connections.GetZone(name, overrides)
		return err
	},
}

func init() {
	var name string
	overrides := false

	GetManagedZonesCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the managedzone")
	GetManagedZonesCmd.Flags().BoolVarP(&overrides, "overrides", "",
		false, "fetch zone details details for use with scaffold")

	_ = GetManagedZonesCmd.MarkFlagRequired("name")
}
