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
	"internal/cmd/utils"
	"os"

	"github.com/spf13/cobra"
)

// CrtTestCaseCmd to get integration flow
var CrtTestCaseCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an integration flow version test case",
	Long:  "Create an integration flow version test case",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := utils.GetStringParam(cmd.Flag("proj"))
		cmdRegion := utils.GetStringParam(cmd.Flag("reg"))
		version := utils.GetStringParam(cmd.Flag("ver"))
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))

		if err = apiclient.SetRegion(cmdRegion); err != nil {
			return err
		}
		if userLabel == "" && version == "" && snapshot == "" {
			return errors.New("at least one of userLabel, version or snapshot must be passed")
		}
		if err = validate(version, userLabel, snapshot, false); err != nil {
			return err
		}

		return apiclient.SetProjectID(cmdProject)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		version := utils.GetStringParam(cmd.Flag("ver"))
		name := utils.GetStringParam(cmd.Flag("name"))
		contentPath := utils.GetStringParam(cmd.Flag("test-case-path"))
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))

		if _, err := os.Stat(contentPath); os.IsNotExist(err) {
			return err
		}

		content, err := os.ReadFile(contentPath)
		if err != nil {
			return err
		}

		if version != "" {
			return apiclient.PrettyPrint(integrations.CreateTestCase(name, version, string(content)))
		} else if userLabel != "" {
			return apiclient.PrettyPrint(integrations.CreateTestCaseByUserLabel(name, userLabel, string(content)))
		} else if snapshot != "" {
			return apiclient.PrettyPrint(integrations.CreateTestCaseBySnapshot(name, snapshot, string(content)))
		}

		return err
	},
}

func init() {
	var name, version, contentPath, userLabel, snapshot string

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
