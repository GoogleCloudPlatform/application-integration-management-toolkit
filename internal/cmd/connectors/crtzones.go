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

package connectors

import (
	"internal/apiclient"
	"internal/client/connections"
	"internal/clilog"
	"internal/cmd/utils"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CreateManagedZonesCmd to list Connections
var CreateManagedZonesCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a managedzone configuration",
	Long:  "Create a managedzone configuration",
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

		description := utils.GetStringParam(cmd.Flag("description"))
		targetProject := utils.GetStringParam(cmd.Flag("target-project"))
		targetVPC := utils.GetStringParam(cmd.Flag("target-vpc"))
		dns := utils.GetStringParam(cmd.Flag("dns"))

		zone := []string{}
		zone = append(zone, "\"dns\":\""+dns+"\"")
		if description != "" {
			zone = append(zone, "\"description\":\""+description+"\"")
		}
		zone = append(zone, "\"targetProject\":\""+targetProject+"\"")
		zone = append(zone, "\"targetVpc\":\""+targetVPC+"\"")

		payload := "{" + strings.Join(zone, ",") + "}"

		return apiclient.PrettyPrint(connections.CreateZone(utils.GetStringParam(cmd.Flag("name")),
			[]byte(payload)))
	},
}

func init() {
	var name, dns, targetProject, targetVPC, description string

	CreateManagedZonesCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the managedzone")
	CreateManagedZonesCmd.Flags().StringVarP(&description, "description", "d",
		"", "Description for the managedzone")
	CreateManagedZonesCmd.Flags().StringVarP(&targetProject, "target-project", "",
		"", "The project where the private DNS zone exists")
	CreateManagedZonesCmd.Flags().StringVarP(&targetVPC, "target-vpc", "",
		"", "The name of the VPC")
	CreateManagedZonesCmd.Flags().StringVarP(&dns, "dns", "",
		"", "DNS name of the resource")

	_ = CreateManagedZonesCmd.MarkFlagRequired("name")
	_ = CreateManagedZonesCmd.MarkFlagRequired("target-project")
	_ = CreateManagedZonesCmd.MarkFlagRequired("target-vpc")
	_ = CreateManagedZonesCmd.MarkFlagRequired("dns")
}
