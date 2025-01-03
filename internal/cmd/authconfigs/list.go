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

package authconfigs

import (
	"internal/apiclient"
	"internal/client/authconfigs"
	"internal/clilog"
	"internal/cmd/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ListCmd to list Integrations
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all integration flows in the region",
	Long:  "List all integration flows in the region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		project := utils.GetStringParam(cmd.Flag("proj"))
		region := utils.GetStringParam(cmd.Flag("reg"))

		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		pageToken := utils.GetStringParam(cmd.Flag("pageToken"))
		filter := utils.GetStringParam(cmd.Flag("filter"))

		_, err = authconfigs.List(pageSize, pageToken, filter)
		return err
	},
}

var pageSize int

func init() {
	var pageToken, filter string

	ListCmd.Flags().IntVarP(&pageSize, "pageSize", "",
		-1, "The maximum number of versions to return")
	ListCmd.Flags().StringVarP(&pageToken, "pageToken", "",
		"", "A page token, received from a previous call")
	ListCmd.Flags().StringVarP(&filter, "filter", "",
		"", "Filter results")
}
