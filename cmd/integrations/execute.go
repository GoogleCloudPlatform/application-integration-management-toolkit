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
	"io/ioutil"
	"os"

	"github.com/srinandan/integrationcli/apiclient"
	"github.com/srinandan/integrationcli/client/integrations"

	"github.com/spf13/cobra"
)

// ExecuteCmd an Integration
var ExecuteCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute an integration",
	Long:  "execute an integration",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		if _, err := os.Stat(executionFile); os.IsNotExist(err) {
			return err
		}

		content, err := ioutil.ReadFile(executionFile)
		if err != nil {
			return err
		}

		_, err = integrations.Execute(name, content)
		return

	},
}

var executionFile string

func init() {
	ExecuteCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ExecuteCmd.Flags().StringVarP(&executionFile, "file", "f",
		"", "Integration payload JSON file path. For the payload structure, visit docs at"+
			" https://cloud.google.com/application-integration/docs/reference/rest/v1/projects.locations.integrations/execute#request-body")

	_ = ExecuteCmd.MarkFlagRequired("name")
	_ = ExecuteCmd.MarkFlagRequired("file")
}
