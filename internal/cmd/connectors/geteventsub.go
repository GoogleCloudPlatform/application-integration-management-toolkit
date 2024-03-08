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

	"github.com/spf13/cobra"
)

// GetEventSubCmd to get connection
var GetEventSubCmd = &cobra.Command{
	Use:   "get",
	Short: "Get event subscription details",
	Long:  "Get event subscription details",
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
		conn := cmd.Flag("conn").Value.String()
		_, err = connections.GetEventSubscription(name, conn, overrides)
		return err
	},
}

func init() {
	var name, conn string

	GetEventSubCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the event subscription")
	GetEventSubCmd.Flags().StringVarP(&conn, "conn", "c",
		"", "The name of the connection")
	GetEventSubCmd.Flags().BoolVarP(&overrides, "overrides", "",
		false, "Provide overrides for the event subscription")

	_ = GetEventSubCmd.MarkFlagRequired("name")
	_ = GetEventSubCmd.MarkFlagRequired("conn")
}
