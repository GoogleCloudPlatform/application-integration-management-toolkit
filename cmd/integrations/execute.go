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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/apiclient"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/client/integrations"

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
		if executionFile != "" && triggerId != "" {
			return fmt.Errorf("cannot pass trigger id and execution file")
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		var content []byte

		if executionFile != "" {
			if _, err := os.Stat(executionFile); os.IsNotExist(err) {
				return err
			}

			content, err = ioutil.ReadFile(executionFile)
			if err != nil {
				return err
			}
		} else if triggerId != "" {
			content = []byte(fmt.Sprintf("{\"triggerId\": \"api_trigger/%s\",\"inputParameters\": {}}", triggerId))
		}

		_, err = integrations.Execute(name, content)
		return

	},
}

var executionFile, triggerId string

func init() {
	ExecuteCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ExecuteCmd.Flags().StringVarP(&executionFile, "file", "f",
		"", "Integration payload JSON file path. For the payload structure, visit docs at"+
			" https://cloud.google.com/application-integration/docs/reference/rest/v1/projects.locations.integrations/execute#request-body")
	ExecuteCmd.Flags().StringVarP(&executionFile, "trigger-id", "",
		"", "Specify only the trigger id of the integration if there are no input parameters to be sent. Cannot be combined with -f")

	_ = ExecuteCmd.MarkFlagRequired("name")
	_ = ExecuteCmd.MarkFlagRequired("file")
}
