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

package connectors

import (
	"internal/apiclient"

	"internal/client/connections"

	"github.com/spf13/cobra"
)

// GetOperationCmd to get connection
var GetOperationCmd = &cobra.Command{
	Use:   "get",
	Short: "Get operation details",
	Long:  "Get operation details from a connection created in a region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		_, err = connections.GetOperation(name)
		return
	},
}

func init() {
	GetOperationCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the operation")

	_ = GetOperationCmd.MarkFlagRequired("name")
}
