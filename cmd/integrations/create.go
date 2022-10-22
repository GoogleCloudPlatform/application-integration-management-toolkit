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
	"io/ioutil"
	"os"

	"github.com/srinandan/integrationcli/apiclient"
	"github.com/srinandan/integrationcli/client/integrations"

	"github.com/spf13/cobra"
)

//CreateCmd to list Integrations
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
		if _, err := os.Stat(filePath); err != nil {
			return err
		}

		content, err := ioutil.ReadFile(integrationFile)
		if err != nil {
			return err
		}
		_, err = integrations.Create(name, content, newIntegration)
		return

	},
}

var integrationFile string
var newIntegration bool

func init() {
	CreateCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	CreateCmd.Flags().StringVarP(&integrationFile, "file", "f",
		"", "Integration flow instance")
	CreateCmd.Flags().BoolVarP(&newIntegration, "new", "",
		false, "Set this flag to true, if draft version is to be created for a brand new integration")

	_ = CreateCmd.MarkFlagRequired("name")
	_ = CreateCmd.MarkFlagRequired("file")
}
