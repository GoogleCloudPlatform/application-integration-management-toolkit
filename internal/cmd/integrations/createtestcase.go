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
	"errors"
	"internal/apiclient"
	"internal/client/integrations"
	"os"

	"github.com/spf13/cobra"
)

// CrtTestCaseCmd to get integration flow
var CrtTestCaseCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an integration flow version test case",
	Long:  "Create an integration flow version test case",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")
		version := cmd.Flag("ver").Value.String()

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		if userLabel == "" && version == "" && snapshot == "" {
			return errors.New("at least one of userLabel, version or snapshot must be passed")
		}
		if err = validate(version); err != nil {
			return err
		}
		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}

		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		version := cmd.Flag("ver").Value.String()
		name := cmd.Flag("name").Value.String()
		contentPath := cmd.Flag("test-case-path").Value.String()

		if _, err := os.Stat(contentPath); os.IsNotExist(err) {
			return err
		}

		content, err := os.ReadFile(contentPath)
		if err != nil {
			return err
		}

		if version != "" {
			_, err = integrations.CreateTestCase(name, version, string(content))
		} else if userLabel != "" {
			_, err = integrations.CreateTestCaseByUserLabel(name, userLabel, string(content))
		} else if snapshot != "" {
			_, err = integrations.CreateTestCaseBySnapshot(name, snapshot, string(content))
		}

		return err
	},
}

func init() {
	var name, version, contentPath string

	CrtTestCaseCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	CrtTestCaseCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	CrtTestCaseCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	CrtTestCaseCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	CrtTestCaseCmd.Flags().StringVarP(&contentPath, "test-case-path", "c",
		"", "Path to a file containing the test case content")

	_ = CrtTestCaseCmd.MarkFlagRequired("name")
	_ = CrtTestCaseCmd.MarkFlagRequired("test-case-path")
}
