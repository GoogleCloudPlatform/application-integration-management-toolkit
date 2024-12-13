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

	"github.com/spf13/cobra"
)

// ListTestCaseCmd to get integration flow
var ListTestCaseCmd = &cobra.Command{
	Use:   "list",
	Short: "List integration flow version test cases",
	Long:  "List integration flow version test cases",
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
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		version := cmd.Flag("ver").Value.String()
		name := cmd.Flag("name").Value.String()

		if version != "" {
			_, err = integrations.ListTestCases(name, version, full)
		} else if userLabel != "" {
			_, err = integrations.ListTestCasesByUserlabel(name, userLabel, full)
		} else if snapshot != "" {
			_, err = integrations.ListTestCasesBySnapshot(name, snapshot, full)
		}

		return err
	},
}

var full bool

func init() {
	var name, version string

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

	_ = ListTestCaseCmd.MarkFlagRequired("name")
}
