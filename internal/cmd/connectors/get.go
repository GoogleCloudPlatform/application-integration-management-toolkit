// Copyright 2021 Google LLC
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
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GetCmd to get connection
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get connection details",
	Long:  "Get connection details from a connection created in a region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(utils.GetStringParam(cmdRegion)); err != nil {
			return err
		}
		if view != "BASIC" && view != "FULL" {
			return fmt.Errorf("view must be BASIC or FULL")
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(utils.GetStringParam(cmdProject))
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		name := utils.GetStringParam(cmd.Flag("name"))
		minimal, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("minimal")))
		overrides, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("overrides")))
		if overrides {
			minimal = true
		}
		_, err = connections.Get(name, view, minimal, overrides)
		return err
	},
}

var view string

func init() {
	var name string
	minimal, overrides := false, false

	GetCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the connection")
	GetCmd.Flags().StringVarP(&view, "view", "",
		"BASIC", "fields of the Connection to be returned; default is BASIC. FULL is the other option")
	GetCmd.Flags().BoolVarP(&minimal, "minimal", "",
		false, "fields of the Connection to be returned; default is false")
	GetCmd.Flags().BoolVarP(&overrides, "overrides", "",
		false, "fetch connector details for use with scaffold")

	_ = GetCmd.MarkFlagRequired("name")
}
