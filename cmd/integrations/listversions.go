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
	"errors"
	"internal/apiclient"

	"internal/client/integrations"

	"github.com/spf13/cobra"
)

// ListVerCmd to list versions of an integration flow
var ListVerCmd = &cobra.Command{
	Use:   "list",
	Short: "List all versions of an integration flow",
	Long:  "List all versions of an integration flow",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		if allVersions && pageSize != -1 {
			return errors.New("allVersions and pageSize cannot be combined")
		}
		if allVersions && pageToken != "" {
			return errors.New("allVersions and pageToken cannot be combined")
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		_, err = integrations.ListVersions(name, pageSize, pageToken, filter, orderBy, false, false, basic)
		return

	},
}

var orderBy string
var basic bool

func init() {
	ListVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ListVerCmd.Flags().IntVarP(&pageSize, "pageSize", "",
		-1, "The maximum number of versions to return")
	ListVerCmd.Flags().StringVarP(&pageToken, "pageToken", "",
		"", "A page token, received from a previous call")
	ListVerCmd.Flags().StringVarP(&filter, "filter", "",
		"", "Filter results")
	ListVerCmd.Flags().StringVarP(&orderBy, "orderBy", "",
		"", "The results would be returned in order")
	ListVerCmd.Flags().BoolVarP(&basic, "basic", "b",
		false, "Returns snapshot and version only")

	_ = ListVerCmd.MarkFlagRequired("name")
}
