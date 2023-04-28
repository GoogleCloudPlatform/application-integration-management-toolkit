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
	"strconv"

	"internal/client/connections"

	"github.com/spf13/cobra"
)

// GetCmd to get endpoint attachments
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an endpoint attachments in the region",
	Long:  "Get an endpoint attachments in the region",
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
		overrides, _ := strconv.ParseBool(cmd.Flag("overrides").Value.String())

		_, err = connections.GetEndpoint(name, overrides)
		return
	},
}

func init() {
	var name string
	var overrides bool

	GetCmd.Flags().StringVarP(&name, "name", "n",
		"", "Endpoint attachment name")
	GetCmd.Flags().BoolVarP(&overrides, "overrides", "",
		false, "Only returns overriable values")

	_ = GetCmd.MarkFlagRequired("name")
}
