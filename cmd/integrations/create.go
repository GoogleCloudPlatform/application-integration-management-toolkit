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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/srinandan/integrationcli/apiclient"
	"github.com/srinandan/integrationcli/client/integrations"

	"github.com/spf13/cobra"
)

// CreateCmd to list Integrations
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an integration flow with a draft version",
	Long:  "Create an integration flow with a draft version",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if _, err := os.Stat(integrationFile); os.IsNotExist(err) {
			fmt.Println(err)
			return err
		}

		content, err := ioutil.ReadFile(integrationFile)
		if err != nil {
			return err
		}
		_, err = integrations.CreateVersion(name, content, snapshot, userLabel)
		return

	},
}

var integrationFile string

func init() {
	CreateCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	CreateCmd.Flags().StringVarP(&integrationFile, "file", "f",
		"", "Integration flow JSON file content")
	CreateCmd.Flags().StringVarP(&snapshot, "snapshot", "s",
		"", "Integration version snapshot number")
	CreateCmd.Flags().StringVarP(&userLabel, "userlabel", "u",
		"", "Integration version userlabel")

	_ = CreateCmd.MarkFlagRequired("name")
	_ = CreateCmd.MarkFlagRequired("file")
}
