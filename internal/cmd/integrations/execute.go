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
	"fmt"
	"internal/apiclient"
	"internal/client/integrations"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// ExecuteCmd an Integration
var ExecuteCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute an integration",
	Long:  "execute an integration",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		if executionFile != "" && triggerID != "" {
			return errors.New("cannot pass trigger id and execution file")
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var content []byte
		name := cmd.Flag("name").Value.String()
		requestID := cmd.Flag("request-id").Value.String()

		if executionFile != "" {
			if _, err := os.Stat(executionFile); os.IsNotExist(err) {
				return err
			}

			content, err = os.ReadFile(executionFile)
			if err != nil {
				return err
			}
		} else if triggerID != "" {
			if requestID == "" {
				requestID = uuid.New().String()
			}
			content = []byte(fmt.Sprintf("{\"triggerId\": \"api_trigger/%s\",\"doNotPropagateError\": %t,\"requestId\": \"%s\",\"inputParameters\": {}}",
				triggerID, doNotPropagateError, requestID))
		}

		_, err = integrations.Execute(name, content)
		return err
	},
}

var (
	executionFile, triggerID string
	doNotPropagateError      bool
)

func init() {
	var name, requestID string

	ExecuteCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ExecuteCmd.Flags().StringVarP(&executionFile, "file", "f",
		"", "Integration payload JSON file path. For the payload structure, visit docs at"+
			" https://cloud.google.com/application-integration/docs/reference/"+
			"rest/v1/projects.locations.integrations/execute#request-body")
	ExecuteCmd.Flags().StringVarP(&triggerID, "trigger-id", "",
		"", "Specify only the trigger id of the integration if there "+
			"are no input parameters to be sent. Cannot be combined with -f")
	ExecuteCmd.Flags().StringVarP(&requestID, "request-id", "",
		"", "This is used to de-dup incoming request")
	ExecuteCmd.Flags().BoolVarP(&doNotPropagateError, "do-not-propagate-error", "",
		false, "Flag to determine how to should propagate errors")

	_ = ExecuteCmd.MarkFlagRequired("name")
}
