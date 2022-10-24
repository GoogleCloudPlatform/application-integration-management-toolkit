// Copyright 2020 Google LLC
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

package preferences

import (
	"github.com/spf13/cobra"
	"github.com/srinandan/integrationcli/apiclient"
)

// Cmd to get org details
var SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set default preferences for integrationcli",
	Long:  "Set default preferences for integrationcli",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.WriteDefaultProject(project); err != nil {
			return err
		}

		if err = apiclient.SetDefaultRegion(region); err != nil {
			return err
		}

		if err = apiclient.SetProxy(proxyUrl); err != nil {
			return err
		}

		if nocheck {
			if err = apiclient.SetNoCheck(nocheck); err != nil {
				return err
			}
		}

		if useapigee {
			if err = apiclient.SetUseApigee(useapigee); err != nil {
				return err
			}
		}

		return nil
	},
}

var project, region, proxyUrl string
var nocheck, useapigee bool

func init() {

	SetCmd.Flags().StringVarP(&project, "proj", "p",
		"", "Integration GCP Project name")

	SetCmd.Flags().StringVarP(&region, "reg", "r",
		"", "Integration region name")

	SetCmd.Flags().StringVarP(&proxyUrl, "proxy", "",
		"", "Use http proxy before contacting the control plane")

	SetCmd.Flags().BoolVarP(&nocheck, "nocheck", "",
		false, "Don't check for newer versions of cmd")

	SetCmd.Flags().BoolVarP(&useapigee, "apigee-integration", "",
		false, "Use Apigee Integration; default is false (Application Integration)")
}
