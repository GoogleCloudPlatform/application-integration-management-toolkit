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
	"internal/apiclient"
	"internal/client/connections"
	"internal/clilog"
	"internal/cmd/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ListEventSubCmd to get connection
var ListEventSubCmd = &cobra.Command{
	Use:   "list",
	Short: "List event subscriptions for a connection",
	Long:  "List event subscriptions for a connection",
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
		name := utils.GetStringParam(cmd.Flag("name"))
		_, err = connections.ListEventSubscriptions(name, pageSize,
			utils.GetStringParam(cmd.Flag("pageToken")), utils.GetStringParam(cmd.Flag("filter")),
			utils.GetStringParam(cmd.Flag("orderBy")))
		return err
	},
}

func init() {
	var name, pageToken, filter, orderBy string

	ListEventSubCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the connection")
	ListEventSubCmd.Flags().IntVarP(&pageSize, "pageSize", "",
		-1, "The maximum number of versions to return")
	ListEventSubCmd.Flags().StringVarP(&pageToken, "pageToken", "",
		"", "A page token, received from a previous call")
	ListEventSubCmd.Flags().StringVarP(&filter, "filter", "",
		"", "Filter condition for list")
	ListEventSubCmd.Flags().StringVarP(&orderBy, "orderBy", "",
		"", "Order by parameters")
}
