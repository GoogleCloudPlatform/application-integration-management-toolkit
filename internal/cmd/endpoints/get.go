// Copyright 2023 Google LLC
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

package endpoints

import (
	"internal/apiclient"
	"internal/client/connections"
	"internal/clilog"
	"internal/cmd/utils"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GetCmd to get endpoint attachments
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an endpoint attachments in the region",
	Long:  "Get an endpoint attachments in the region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		project := utils.GetStringParam(cmd.Flag("proj"))
		region := utils.GetStringParam(cmd.Flag("reg"))

		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		name := utils.GetStringParam(cmd.Flag("name"))
		overrides, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("overrides")))

		return apiclient.PrettyPrint(connections.GetEndpoint(name, overrides))
	},
}

func init() {
	var name string
	var overrides bool

	GetCmd.Flags().StringVarP(&name, "name", "n",
		"", "Endpoint attachment name")
	GetCmd.Flags().BoolVarP(&overrides, "overrides", "",
		false, "Only returns overriable values")

	_ = GetCmd.MarkFlagRequired("name")
}
