// Copyright 2022 Google LLC
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

package connectors

import (
	"internal/apiclient"

	"internal/client/connections"

	"github.com/spf13/cobra"
)

// SetCustomCmd to set admin role
var SetCustomCmd = &cobra.Command{
	Use:   "setcustom",
	Short: "Set a custom IAM role on a Connection",
	Long:  "Set a custom IAM role on a Connection",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := cmd.Flag("name").Value.String()
		return connections.SetIAM(name, memberName, role, memberType)
	},
}

func init() {
	var name string

	SetCustomCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the connection")
	SetCustomCmd.Flags().StringVarP(&memberName, "member", "m",
		"", "Member Name, example Service Account Name")
	SetCustomCmd.Flags().StringVarP(&role, "role", "",
		"", "Custom IAM role in the format projects/{project-id}/roles/{role}")
	SetCustomCmd.Flags().StringVarP(&memberType, "member-type", "",
		"serviceAccount", "memberType must be serviceAccount, user, or group (default serviceAccount)")

	_ = SetCustomCmd.MarkFlagRequired("name")
	_ = SetCustomCmd.MarkFlagRequired("role")
}
