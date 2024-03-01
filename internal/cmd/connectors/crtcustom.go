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

	"github.com/spf13/cobra"
)

// CrtCustomCmd to create a new connection
var CrtCustomCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new custom connection",
	Long:  "Create a new customer connection in a region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		connType := cmd.Flag("type").Value.String()
		if connType != "OPEN_API" && connType != "PROTO" {
			return fmt.Errorf("connection type must be OPEN_API or PROTO")
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := cmd.Flag("name").Value.String()
		description := cmd.Flag("description").Value.String()
		displayName := cmd.Flag("display-name").Value.String()
		connType := cmd.Flag("type").Value.String()

		_, err = connections.CreateCustom(name, description, displayName, connType, labels)

		return err
	},
}

var labels map[string]string

func init() {
	var name, description, displayName, connType string

	CrtCustomCmd.Flags().StringVarP(&name, "name", "n",
		"", "Connection name")
	CrtCustomCmd.Flags().StringVarP(&displayName, "display-name", "d",
		"", "Custom Connection display name")
	CrtCustomCmd.Flags().StringVarP(&description, "description", "",
		"", "Custom Connection description")
	CrtCustomCmd.Flags().StringVarP(&connType, "type", "",
		"", "Custom Connection type")
	CrtCustomCmd.Flags().StringToStringVarP(&labels, "labels", "l",
		map[string]string{}, "Custom Connection labels")

	_ = CrtCustomCmd.MarkFlagRequired("name")
}
