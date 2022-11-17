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
	"io/ioutil"
	"os"

	"github.com/srinandan/integrationcli/apiclient"
	"github.com/srinandan/integrationcli/client/connections"

	"github.com/spf13/cobra"
)

// CreateCmd to create a new connection
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new connection",
	Long:  "Create a new connection in a region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if _, err := os.Stat(connectionFile); os.IsNotExist(err) {
			return err
		}

		content, err := ioutil.ReadFile(connectionFile)
		if err != nil {
			return err
		}

		_, err = connections.Create(name, content, grantPermission)
		return
	},
}

var connectionFile string
var grantPermission bool

func init() {
	CreateCmd.Flags().StringVarP(&name, "name", "n",
		"", "Connection name")
	CreateCmd.Flags().StringVarP(&connectionFile, "file", "f",
		"", "Connection details JSON file path")
	CreateCmd.Flags().BoolVarP(&grantPermission, "grant-permission", "g",
		false, "Grant the service account permission to the GCP resource")
}
