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
	"internal/clilog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ListManagedZonesCmd to list Connections
var ListManagedZonesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managedzones configured",
	Long:  "List all managedzones configured",
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
		_, err = connections.ListZones(pageSize,
			cmd.Flag("pageToken").Value.String(),
			cmd.Flag("filter").Value.String(),
			cmd.Flag("orderBy").Value.String())
		return err
	},
}

func init() {
	var pageToken, filter, orderBy string

	ListManagedZonesCmd.Flags().IntVarP(&pageSize, "pageSize", "",
		-1, "The maximum number of versions to return")
	ListManagedZonesCmd.Flags().StringVarP(&pageToken, "pageToken", "",
		"", "A page token, received from a previous call")
	ListManagedZonesCmd.Flags().StringVarP(&filter, "filter", "",
		"", "Filter results")
	ListManagedZonesCmd.Flags().StringVarP(&orderBy, "orderBy", "",
		"", "The results would be returned in order")
}
