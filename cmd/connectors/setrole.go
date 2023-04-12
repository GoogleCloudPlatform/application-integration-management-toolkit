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
	"errors"
	"internal/apiclient"

	"internal/client/connections"

	"github.com/spf13/cobra"
)

// SetRoleCmd to set admin role
var SetRoleCmd = &cobra.Command{
	Use:   "setrole",
	Short: "Set Connection IAM policy on a Connection",
	Long:  "Set Connection IAM policy on a Connection",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(cmd.Flag("reg").Value.String()); err != nil {
			return err
		}
		return apiclient.SetProjectID(cmd.Flag("proj").Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := cmd.Flag("name").Value.String()
		role := cmd.Flag("role").Value.String()
		memberName := cmd.Flag("member").Value.String()
		memberType := cmd.Flag("member-type").Value.String()
		return connections.SetIAM(name, memberName, role, memberType)
	},
}

type connectorRole string

const (
	admin   connectorRole = "admin"
	custom  connectorRole = "custom"
	invoker connectorRole = "invoker"
	viewer  connectorRole = "viewer"
)

func (r *connectorRole) String() string {
	return string(*r)
}

func (c *connectorRole) Set(r string) error {
	switch r {
	case "admin", "custom", "invoker", "viewer":
		*c = connectorRole(r)
		return nil
	default:
		return errors.New(`must be one of "admin", "custom", "viewer" or "invoker"`)
	}
}

func (e *connectorRole) Type() string {
	return "connectorRole"
}

func init() {
	var memberName, memberType, name string
	var role connectorRole

	SetRoleCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the connection")
	SetRoleCmd.Flags().Var(&role, "role", "The name of the role, must be admin, invoker, custom or viewer")
	SetRoleCmd.Flags().StringVarP(&memberName, "member", "m",
		"", "Member Name, example Service Account Name")
	SetRoleCmd.Flags().StringVarP(&memberType, "member-type", "",
		"serviceAccount", "memberType must be serviceAccount, user, or group (default serviceAccount)")

	_ = SetRoleCmd.MarkFlagRequired("name")
	_ = SetRoleCmd.MarkFlagRequired("role")
}
