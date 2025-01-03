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
	"internal/apiclient"
	"internal/client/connections"
	"internal/clilog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GetCustomVerCmd to get connection
var GetCustomVerCmd = &cobra.Command{
	Use:   "get",
	Short: "Get custom connection version details",
	Long:  "Get custom connection version details from a connection created in a region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := cmd.Flag("name").Value.String()
		id := cmd.Flag("id").Value.String()
		_, err = connections.GetCustomVersion(name, id, overrides)
		return err
	},
}

var overrides bool

func init() {
	var name, id string

	GetCustomVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the custom connection")
	GetCustomVerCmd.Flags().StringVarP(&id, "id", "",
		"", "Identifier assigned to the custom connection version")
	GetCustomVerCmd.Flags().BoolVarP(&overrides, "overrides", "",
		false, "Provide overrides for the custom connection version")

	_ = GetCustomVerCmd.MarkFlagRequired("name")
	_ = GetCustomVerCmd.MarkFlagRequired("id")
}
