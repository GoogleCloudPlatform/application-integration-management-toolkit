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
	"fmt"
	"internal/apiclient"
	"internal/client/connections"
	"internal/clilog"
	"internal/cmd/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// SetRoleCmd to set admin role
var SetRoleCmd = &cobra.Command{
	Use:   "setrole",
	Short: "Set Connection IAM policy on a Connection",
	Long:  "Set Connection IAM policy on a Connection",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(utils.GetStringParam(cmd.Flag("reg"))); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(utils.GetStringParam(cmd.Flag("proj")))
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		name := utils.GetStringParam(cmd.Flag("name"))
		role := utils.GetStringParam(cmd.Flag("role"))
		memberName := utils.GetStringParam(cmd.Flag("member"))
		memberType := utils.GetStringParam(cmd.Flag("member-type"))
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

func (c *connectorRole) String() string {
	return string(*c)
}

func (c *connectorRole) Set(r string) error {
	switch r {
	case "admin", "custom", "invoker", "viewer":
		*c = connectorRole(r)
		return nil
	default:
		return fmt.Errorf("must be one of %s,%s, %s or %s", admin, custom, invoker, viewer)
	}
}

func (c *connectorRole) Type() string {
	return "connectorRole"
}

func init() {
	var memberName, memberType, name string
	var role connectorRole

	SetRoleCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the connection")
	SetRoleCmd.Flags().Var(&role, "role",
		fmt.Sprintf("The role must be one of %s,%s, %s or %s", admin, custom, invoker, viewer))
	SetRoleCmd.Flags().StringVarP(&memberName, "member", "m",
		"", "Member Name, example Service Account Name")
	SetRoleCmd.Flags().StringVarP(&memberType, "member-type", "",
		"serviceAccount", "memberType must be serviceAccount, user, or group (default serviceAccount)")

	_ = SetRoleCmd.MarkFlagRequired("name")
	_ = SetRoleCmd.MarkFlagRequired("role")
}
