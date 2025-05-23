// Copyright 2024 Google LLC
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
	"internal/apiclient"
	"internal/client/integrations"

	"github.com/spf13/cobra"
)

// DelTestCaseCmd to get integration flow
var DelTestCaseCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes an integration flow version test case",
	Long:  "Deletes an integration flow version test case",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}

		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		version := cmd.Flag("ver").Value.String()
		name := cmd.Flag("name").Value.String()
		testCaseID := cmd.Flag("test-case-id").Value.String()
		_, err = integrations.DeleteTestCase(name, version, testCaseID)
		return err
	},
}

func init() {
	var name, version, testCaseID string

	DelTestCaseCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	DelTestCaseCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	DelTestCaseCmd.Flags().StringVarP(&testCaseID, "test-case-id", "c",
		"", "Test Case ID")
	_ = DelTestCaseCmd.MarkFlagRequired("name")
	_ = DelTestCaseCmd.MarkFlagRequired("ver")
	_ = DelTestCaseCmd.MarkFlagRequired("test-case-id")

}
