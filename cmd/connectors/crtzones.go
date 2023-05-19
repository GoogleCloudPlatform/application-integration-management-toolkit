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
	"internal/apiclient"
	"strings"

	"internal/client/connections"

	"github.com/spf13/cobra"
)

// CreateManagedZonesCmd to list Connections
var CreateManagedZonesCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a managedzone configuration",
	Long:  "Create a managedzone configuration",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		description := cmd.Flag("description").Value.String()
		targetProject := cmd.Flag("target-project").Value.String()
		targetVPC := cmd.Flag("target-vpc").Value.String()
		dns := cmd.Flag("dns").Value.String()

		zone := []string{}
		zone = append(zone, "\"dns\":\""+dns+"\"")
		if description != "" {
			zone = append(zone, "\"description\":\""+description+"\"")
		}
		zone = append(zone, "\"targetProject\":\""+targetProject+"\"")
		zone = append(zone, "\"targetVpc\":\""+targetVPC+"\"")

		payload := "{" + strings.Join(zone, ",") + "}"

		_, err = connections.CreateZone(cmd.Flag("name").Value.String(),
			[]byte(payload))
		return
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
