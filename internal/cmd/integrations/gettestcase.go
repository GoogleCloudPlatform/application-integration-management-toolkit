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
	"strconv"

	"github.com/spf13/cobra"
)

// GetTestCaseCmd to get integration flow
var GetTestCaseCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an integration flow version test case",
	Long:  "Get an integration flow version test case",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := utils.GetStringParam(cmd.Flag("proj"))
		cmdRegion := utils.GetStringParam(cmd.Flag("reg"))
		version := utils.GetStringParam(cmd.Flag("ver"))
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))
		latest, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("latest")))

		if err = apiclient.SetRegion(cmdRegion); err != nil {
			return err
		}
		if err = validate(version, userLabel, snapshot, latest); err != nil {
			return err
		}

		return apiclient.SetProjectID(cmdProject)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		var integrationBody []byte

		version := utils.GetStringParam(cmd.Flag("ver"))
		name := utils.GetStringParam(cmd.Flag("name"))
		testCaseID := utils.GetStringParam(cmd.Flag("test-case-id"))
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))

		apiclient.DisableCmdPrintHttpResponse()

		latest := ignoreLatest(version, userLabel, snapshot)
		if latest {
			if version, err = getLatestVersion(name); err != nil {
				return err
			}
		} else {
			if version != "" {
				integrationBody, err = integrations.Get(name, version, true, false, false)
			} else if snapshot != "" {
				integrationBody, err = integrations.GetBySnapshot(name, snapshot, true, false, false)
			} else if userLabel != "" {
				integrationBody, err = integrations.GetByUserlabel(name, userLabel, true, false, false)
			} else {
				return errors.New("latest version not found. Must pass oneOf version, snapshot or user-label or fix the integration name")
			}
			version, err = getIntegrationVersion(integrationBody)
			if err != nil {
				return err
			}
		}

		apiclient.EnableCmdPrintHttpResponse()

		_, err = integrations.GetTestCase(name, version, testCaseID)
		return err
	},
}

func init() {
	var name, version, testCaseID, userLabel, snapshot string
	var latest bool

	GetTestCaseCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	GetTestCaseCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	GetTestCaseCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	GetTestCaseCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	GetTestCaseCmd.Flags().BoolVarP(&latest, "latest", "",
		true, "Get the version with the highest snapshot number in SNAPSHOT state. If none found, selects the highest snapshot in DRAFT state; default is true")
	GetTestCaseCmd.Flags().StringVarP(&testCaseID, "test-case-id", "c",
		"", "Test Case ID")
	_ = GetTestCaseCmd.MarkFlagRequired("name")
	_ = GetTestCaseCmd.MarkFlagRequired("ver")
	_ = GetTestCaseCmd.MarkFlagRequired("test-case-id")

}
