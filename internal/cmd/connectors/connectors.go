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

	"github.com/spf13/cobra"
)

// Cmd to manage preferences
var Cmd = &cobra.Command{
	Use:   "connectors",
	Short: "Manage connections for Integration Connectors",
	Long:  "Manage connections for Integration Connectors",
}

var examples = []string{
	`integrationcli connectors create -n $name -g=true -f samples/pub_sub_connection.json -sa=connectors --wait=true --default-token`,
	`integrationcli connectors create -n $name -f samples/gcs_connection.json -sa=connectors --wait=true --default-token`,
	`integrationcli connectors custom versions create --id $version -n $name -f samples/custom-connection.json --sa=connectors --default-token`,
	`integrationcli connectors custom create -n $name -d $dispName --type OPEN_API --default-token`,
}

type ConnectorType string

const (
	OPEN_API ConnectorType = "OPEN_API"
	PROTO    ConnectorType = "PROTO"
)

func init() {
	var region, project string

	Cmd.PersistentFlags().StringVarP(&project, "proj", "p",
		"", "Integration GCP Project name")

	Cmd.PersistentFlags().StringVarP(&region, "reg", "r",
		"", "Integration region name")

	Cmd.AddCommand(CreateCmd)
	Cmd.AddCommand(DelCmd)
	Cmd.AddCommand(ListCmd)
	Cmd.AddCommand(GetCmd)
	Cmd.AddCommand(IamCmd)
	Cmd.AddCommand(NodeCountCmd)
	Cmd.AddCommand(ExportCmd)
	Cmd.AddCommand(ImportCmd)
	Cmd.AddCommand(PatchCmd)
	Cmd.AddCommand(OperationsCmd)
	Cmd.AddCommand(ManagedZonesCmd)
	Cmd.AddCommand(CustomCmd)
	Cmd.AddCommand(EventSubCmd)
}

func GetExample(i int) string {
	return examples[i]
}

func (c *ConnectorType) String() string {
	return string(*c)
}

func (c *ConnectorType) Set(r string) error {
	switch r {
	case "OPEN_API", "PROTO":
		*c = ConnectorType(r)
	default:
		return fmt.Errorf("must be %s or %s", OPEN_API, PROTO)
	}
	return nil
}

func (c *ConnectorType) Type() string {
	return "ConnectorType"
}
