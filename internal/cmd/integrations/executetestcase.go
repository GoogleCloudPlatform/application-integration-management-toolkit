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
	"internal/cmd/utils"
	"os"

	"github.com/spf13/cobra"
)

// ExecuteTestCaseCmd to get integration flow
var ExecuteTestCaseCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute an integration flow version test case",
	Long:  "Execute an integration flow version test case",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := utils.GetStringParam(cmd.Flag("proj"))
		cmdRegion := utils.GetStringParam(cmd.Flag("reg"))

		if err = apiclient.SetRegion(cmdRegion); err != nil {
			return err
		}

		return apiclient.SetProjectID(cmdProject)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		version := utils.GetStringParam(cmd.Flag("ver"))
		name := utils.GetStringParam(cmd.Flag("name"))
		testCaseID := utils.GetStringParam(cmd.Flag("test-case-id"))
		inputFile := utils.GetStringParam(cmd.Flag("input-file"))

		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return err
		}

		content, err := os.ReadFile(inputFile)
		if err != nil {
			return err
		}

		_, err = integrations.ExecuteTestCase(name, version, testCaseID, string(content))
		return err
	},
}

func init() {
	var name, version, testCaseID, inputFile string

	ExecuteTestCaseCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ExecuteTestCaseCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	ExecuteTestCaseCmd.Flags().StringVarP(&testCaseID, "test-case-id", "c",
		"", "Test Case ID")
	ExecuteTestCaseCmd.Flags().StringVarP(&inputFile, "input-file", "f",
		"", "Path to a file containing input parameters")
	_ = ExecuteTestCaseCmd.MarkFlagRequired("name")
	_ = ExecuteTestCaseCmd.MarkFlagRequired("ver")
	_ = ExecuteTestCaseCmd.MarkFlagRequired("test-case-id")

}
