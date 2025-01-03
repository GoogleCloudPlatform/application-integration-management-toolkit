// Copyright 2022 Google LLC
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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// DelCmd to delete integrations
var DelCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes an Integration and all versions of it",
	Long:  "Deletes an Integration and all versions of it",
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
		_, err = integrations.Delete(name)
		return err
	},
}

func init() {
	var name string

	DelCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration name")

	_ = DelCmd.MarkFlagRequired("name")
}
