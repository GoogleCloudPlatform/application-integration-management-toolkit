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

package endpoints

import (
	"internal/apiclient"
	"internal/client/connections"

	"github.com/spf13/cobra"
)

// ListCmd to list Integrations
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all endpoint attachments in the region",
	Long:  "List all endpoint attachments in the region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		project := cmd.Flag("proj").Value.String()
		region := cmd.Flag("reg").Value.String()

		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		pageToken := cmd.Flag("pageToken").Value.String()
		filter := cmd.Flag("filter").Value.String()

		_, err = connections.ListEndpoints(pageSize, pageToken, filter, "")
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
