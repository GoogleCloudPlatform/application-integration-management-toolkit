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

	"internal/apiclient"

	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/client/integrations"

	"github.com/spf13/cobra"
)

// GetVerCmd to get integration flow
var GetVerCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an integration flow version",
	Long:  "Get an integration flow version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		if snapshot == "" && userLabel == "" && version == "" {
			return fmt.Errorf("at least one of snapshot, userLabel and version must be supplied")
		}
		if snapshot != "" && (userLabel != "" || version != "") {
			return fmt.Errorf("snapshot cannot be combined with userLabel or version")
		}
		if userLabel != "" && (snapshot != "" || version != "") {
			return fmt.Errorf("userLabel cannot be combined with snapshot or version")
		}
		if version != "" && (snapshot != "" || userLabel != "") {
			return fmt.Errorf("version cannot be combined with snapshot or version")
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if version != "" {
			_, err = integrations.Get(name, version, basic, minimal, overrides)
		} else if snapshot != "" {
			_, err = integrations.GetBySnapshot(name, snapshot, minimal, overrides)
		} else {
			_, err = integrations.GetByUserlabel(name, userLabel, minimal, overrides)
		}
		return

	},
}

var minimal bool

func init() {
	GetVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	GetVerCmd.Flags().StringVarP(&version, "ver", "v",
		"", "Integration flow version")
	GetVerCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration flow snapshot number")
	GetVerCmd.Flags().StringVarP(&userLabel, "user-label", "u",
		"", "Integration flow user label")
	GetVerCmd.Flags().BoolVarP(&basic, "basic", "b",
		false, "Returns snapshot and version only")
	GetVerCmd.Flags().BoolVarP(&overrides, "overrides", "o",
		false, "Returns overrides only for integration")
	GetVerCmd.Flags().BoolVarP(&minimal, "minimal", "",
		false, "fields of the Integration to be returned; default is false")

	_ = GetVerCmd.MarkFlagRequired("name")
}
