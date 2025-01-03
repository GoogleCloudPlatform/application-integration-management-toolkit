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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ListSuspCmd to list suspensions of an integration
var ListSuspCmd = &cobra.Command{
	Use:   "list",
	Short: "List all suspensions of an integration",
	Long:  "List all suspensions of an integration",
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
		name := cmd.Flag("name").Value.String()
		_, err = integrations.ListSuspensions(name, execution, pageSize,
			cmd.Flag("pageToken").Value.String(),
			cmd.Flag("filter").Value.String(),
			cmd.Flag("orderBy").Value.String())
		return err
	},
}

func init() {
	var name, pageToken, filter, orderBy string

	ListSuspCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ListSuspCmd.Flags().StringVarP(&execution, "execution", "e",
		"", "Execution Id of the integration")
	ListSuspCmd.Flags().IntVarP(&pageSize, "pageSize", "",
		-1, "The maximum number of versions to return")
	ListSuspCmd.Flags().StringVarP(&pageToken, "pageToken", "",
		"", "A page token, received from a previous call")
	ListSuspCmd.Flags().StringVarP(&filter, "filter", "",
		"", "Filter results")
	ListSuspCmd.Flags().StringVarP(&orderBy, "orderBy", "",
		"", "The results would be returned in order")

	_ = ListSuspCmd.MarkFlagRequired("name")
	_ = ListSuspCmd.MarkFlagRequired("execution")
}
