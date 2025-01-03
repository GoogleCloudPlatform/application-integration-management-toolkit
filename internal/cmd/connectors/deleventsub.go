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
	"internal/cmd/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// DelEventSubCmd to get connection
var DelEventSubCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete event subscription",
	Long:  "Delte event subscription",
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
		name := utils.GetStringParam(cmd.Flag("name"))
		conn := utils.GetStringParam(cmd.Flag("conn"))
		_, err = connections.DeleteEventSubscription(name, conn)
		return err
	},
}

func init() {
	var name, conn string

	DelEventSubCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the event subscription")
	DelEventSubCmd.Flags().StringVarP(&conn, "conn", "c",
		"", "The name of the connection")

	_ = DelEventSubCmd.MarkFlagRequired("name")
	_ = DelEventSubCmd.MarkFlagRequired("conn")
}
