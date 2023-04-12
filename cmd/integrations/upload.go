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
	"os"

	"internal/apiclient"

	"internal/client/integrations"

	"github.com/spf13/cobra"
)

// UploadCmd to upload integrations
var UploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an Integration flow",
	Long:  "Upload an Integration flow",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if _, err := os.Stat(filePath); err != nil {
			return err
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		_, err = integrations.Upload(name, content)

		return err
	},
}

var filePath string

func init() {
	UploadCmd.Flags().StringVarP(&name, "name", "n",
		"", "File containing Integration flow name")
	UploadCmd.Flags().StringVarP(&filePath, "file", "f",
		"", "File containing an Integration flow json in stringified format. "+
			"To send json, see integrationcli integrations versions patch")

	_ = UploadCmd.MarkFlagRequired("file")
}
