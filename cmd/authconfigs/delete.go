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

	"github.com/spf13/cobra"
)

// DelCmd to get integration flow
var DelCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an authconfig from a region",
	Long:  "Delete an authconfig from a region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		project := cmd.Flag("proj").Value.String()
		region := cmd.Flag("reg").Value.String()

		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := cmd.Flag("name").Value.String()
		_, err = authconfigs.Delete(name)
		return
	},
}

func init() {
	var name string

	DelCmd.Flags().StringVarP(&name, "name", "n",
		"", "AuthConfig name")

	_ = DelCmd.MarkFlagRequired("name")
}
