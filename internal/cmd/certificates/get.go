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

package certificates

import (
	"errors"
	"internal/apiclient"
	"internal/client/certificates"
	"internal/clilog"
	"internal/cmd/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GetCmd to get integration flow
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get certificate details from a region",
	Long:  "Get certificate details from a region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		project := utils.GetStringParam(cmd.Flag("proj"))
		region := utils.GetStringParam(cmd.Flag("reg"))
		name := utils.GetStringParam(cmd.Flag("name"))
		id := utils.GetStringParam(cmd.Flag("id"))

		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		if id == "" && name == "" {
			return errors.New("id and name cannot be empty")
		}
		if id != "" && name != "" {
			return errors.New("id and name both cannot be set")
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		name := utils.GetStringParam(cmd.Flag("name"))
		id := utils.GetStringParam(cmd.Flag("id"))

		if name != "" {
			version, err := certificates.Find(name)
			if err != nil {
				return err
			}
			return apiclient.PrettyPrint(certificates.Get(version))
		}

		return apiclient.PrettyPrint(certificates.Get(id))
	},
}

func init() {
	var name, id string
	GetCmd.Flags().StringVarP(&id, "id", "i",
		"", "Certificate name (uuid)")
	GetCmd.Flags().StringVarP(&name, "name", "n",
		"", "Certificate display name")
}
