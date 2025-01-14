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

	"github.com/spf13/cobra"
)

// ListTestCaseCmd to get integration flow
var ListTestCaseCmd = &cobra.Command{
	Use:   "list",
	Short: "List integration flow version test cases",
	Long:  "List integration flow version test cases",
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
		userLabel := utils.GetStringParam(cmd.Flag("user-label"))
		snapshot := utils.GetStringParam(cmd.Flag("snapshot"))
		pageToken := utils.GetStringParam(cmd.Flag("pageToken"))
		filter := utils.GetStringParam(cmd.Flag("filter"))
		orderBy := utils.GetStringParam(cmd.Flag("orderBy"))

		if version != "" {
			_, err = integrations.ListTestCases(name, version, full, filter, pageSize, pageToken, orderBy)
		} else if userLabel != "" {
			_, err = integrations.ListTestCasesByUserlabel(name, userLabel, full, filter, pageSize, pageToken, orderBy)
		} else if snapshot != "" {
			_, err = integrations.ListTestCasesBySnapshot(name, snapshot, full, filter, pageSize, pageToken, orderBy)
		}

		return err
	},
}

var full bool

func init() {
	var name, version, userLabel, snapshot, pageToken, filter, orderBy string

	ListTestCaseCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ListTestCaseCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	ListTestCaseCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	ListTestCaseCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	ListTestCaseCmd.Flags().BoolVarP(&full, "full", "",
		false, "Full test case response")
	ListTestCaseCmd.Flags().IntVarP(&pageSize, "pageSize", "",
		-1, "The maximum number of versions to return")
	ListTestCaseCmd.Flags().StringVarP(&pageToken, "pageToken", "",
		"", "A page token, received from a previous call")
	ListTestCaseCmd.Flags().StringVarP(&filter, "filter", "",
		"", "Filter results")
	ListTestCaseCmd.Flags().StringVarP(&orderBy, "orderBy", "",
		"", "The results would be returned in order")

	_ = ListTestCaseCmd.MarkFlagRequired("name")
}
