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

package integrations

import (
	"internal/apiclient"
	"internal/client/integrations"
	"internal/clilog"
	"internal/cmd/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ListVerCmd to list versions of an integration flow
var ListVerCmd = &cobra.Command{
	Use:   "list",
	Short: "List all versions of an integration flow",
	Long:  "List all versions of an integration flow",
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
		cmd.SilenceUsage = true

		name := utils.GetStringParam(cmd.Flag("name"))
		basic := utils.GetBasicInfo(cmd, "basic")
		return apiclient.PrettyPrint(integrations.ListVersions(name, pageSize,
			utils.GetStringParam(cmd.Flag("pageToken")),
			utils.GetStringParam(cmd.Flag("filter")),
			utils.GetStringParam(cmd.Flag("orderBy")),
			false, false, basic))
	},
	Example: `Return a list of versions with basic information: ` + GetExample(3) + `
Return the version that is published: ` + GetExample(4),
}

func init() {
	var pageToken, filter, orderBy, name, basic string

	ListVerCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ListVerCmd.Flags().IntVarP(&pageSize, "pageSize", "",
		-1, "The maximum number of versions to return")
	ListVerCmd.Flags().StringVarP(&pageToken, "pageToken", "",
		"", "A page token, received from a previous call")
	ListVerCmd.Flags().StringVarP(&filter, "filter", "",
		"", "Filter results")
	ListVerCmd.Flags().StringVarP(&orderBy, "orderBy", "",
		"", "The results would be returned in order")
	ListVerCmd.Flags().StringVarP(&basic, "basic", "b",
		"", "Returns snapshot and version only; default is false")

	_ = ListVerCmd.MarkFlagRequired("name")
}
