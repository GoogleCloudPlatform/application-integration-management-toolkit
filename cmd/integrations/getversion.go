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
	"github.com/srinandan/integrationcli/apiclient"
	"github.com/srinandan/integrationcli/client/integrations"

	"github.com/spf13/cobra"
)

// GetVerCmd to get integration flow
var GetVerCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an integration flow version",
	Long:  "Get an integration flow version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		_, err = integrations.Get(name, version, basic)
		return

	},
}

func init() {
	GetVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	GetVerCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	GetVerCmd.Flags().BoolVarP(&basic, "basic", "b",
		false, "Returns snapshot and version only")

	_ = GetVerCmd.MarkFlagRequired("name")
	_ = GetVerCmd.MarkFlagRequired("ver")
}
