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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CrtEventSubCmd to create a new connection
var CrtEventSubCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new event subscription",
	Long:  "Create a new event subscription in a region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(utils.GetStringParam(cmdRegion)); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(utils.GetStringParam(cmdProject))
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := utils.GetStringParam(cmd.Flag("name"))
		id := utils.GetStringParam(cmd.Flag("id"))
		eventsubFile := utils.GetStringParam(cmd.Flag("file"))

		if _, err = os.Stat(eventsubFile); err != nil {
			return fmt.Errorf("unable to open file %w", err)
		}

		contents, err := os.ReadFile(eventsubFile)
		if err != nil {
			return fmt.Errorf("unable to open file %w", err)
		}
		_, err = connections.CreateEventSubscription(name, id, contents)
		return err
	},
}

func init() {
	var name, id, eventsubFile string

	CrtEventSubCmd.Flags().StringVarP(&name, "name", "n",
		"", "Connection name")
	CrtEventSubCmd.Flags().StringVarP(&id, "id", "",
		"", "Identifier to assign to the Event Subscription")
	CrtEventSubCmd.Flags().StringVarP(&eventsubFile, "file", "f",
		"", "Event Subscription details JSON file path")

	_ = CrtEventSubCmd.MarkFlagRequired("name")
	_ = CrtEventSubCmd.MarkFlagRequired("id")
}
