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
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// RepairCmd to repair events of a connection
var RepairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Tries to repair eventing related event subscriptions",
	Long:  "tries to repair eventing related event subscriptions",
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
		wait, _ := strconv.ParseBool(cmd.Flag("wait").Value.String())

		err = connections.RepairEvent(name, wait)
		return err
	},
}

func init() {
	var name string
	var wait bool

	RepairCmd.Flags().StringVarP(&name, "name", "n",
		"", "Connection name")

	RepairCmd.Flags().BoolVarP(&wait, "wait", "",
		false, "Waits for the repair to finish, with success or error; default is false")

	_ = RepairCmd.MarkFlagRequired("updateMask")
}
