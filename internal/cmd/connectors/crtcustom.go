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

// CrtCustomCmd to create a new connection
var CrtCustomCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new custom connection",
	Long:  "Create a new customer connection in a region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(utils.GetStringParam(cmdRegion)); err != nil {
			return err
		}
		connType := utils.GetStringParam(cmd.Flag("type"))
		if connType != "OPEN_API" && connType != "PROTO" {
			return fmt.Errorf("connection type must be OPEN_API or PROTO")
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(utils.GetStringParam(cmdProject))
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		name := utils.GetStringParam(cmd.Flag("name"))
		description := utils.GetStringParam(cmd.Flag("description"))
		displayName := utils.GetStringParam(cmd.Flag("display-name"))
		connType := utils.GetStringParam(cmd.Flag("type"))

		return apiclient.PrettyPrint(connections.CreateCustom(name, description, displayName, connType, labels))
	},
	Example: `Create a custom connector for OPEN_API type: ` + GetExample(3),
}

var (
	labels   map[string]string
	connType ConnectorType
)

func init() {
	var name, description, displayName string

	CrtCustomCmd.Flags().StringVarP(&name, "name", "n",
		"", "Connection name")
	CrtCustomCmd.Flags().StringVarP(&displayName, "display-name", "d",
		"", "Custom Connection display name")
	CrtCustomCmd.Flags().StringVarP(&description, "description", "",
		"", "Custom Connection description")
	CrtCustomCmd.Flags().Var(&connType, "type",
		"Custom Connection type must be set to OPEN_API or PROTO")
	CrtCustomCmd.Flags().StringToStringVarP(&labels, "labels", "l",
		map[string]string{}, "Custom Connection labels")

	_ = CrtCustomCmd.MarkFlagRequired("name")
}
