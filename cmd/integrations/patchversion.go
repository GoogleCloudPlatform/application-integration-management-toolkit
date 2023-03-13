// Copyright 2022 Google LLC
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
	"io/ioutil"
	"os"

	"internal/apiclient"

	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/client/integrations"

	"github.com/spf13/cobra"
)

// PatchVerCmd to get integration flow
var PatchVerCmd = &cobra.Command{
	Use:   "patch",
	Short: "Patch an integration flow version",
	Long:  "Patch an integration flow version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if _, err := os.Stat(integrationFile); os.IsNotExist(err) {
			return err
		}

		content, err := ioutil.ReadFile(integrationFile)
		if err != nil {
			return err
		}
		_, err = integrations.Patch(name, version, content)
		return

	},
}

func init() {
	PatchVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	PatchVerCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	PatchVerCmd.Flags().StringVarP(&integrationFile, "file", "f",
		"", "Integration flow JSON file content")

	_ = PatchVerCmd.MarkFlagRequired("name")
	_ = PatchVerCmd.MarkFlagRequired("ver")
	_ = PatchVerCmd.MarkFlagRequired("file")
}
