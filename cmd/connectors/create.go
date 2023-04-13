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
	"os"
	"regexp"

	"internal/apiclient"

	"internal/client/connections"

	"github.com/spf13/cobra"
)

// CreateCmd to create a new connection
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new connection",
	Long:  "Create a new connection in a region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := cmd.Flag("name").Value.String()
		if _, err := os.Stat(connectionFile); os.IsNotExist(err) {
			return err
		}

		content, err := os.ReadFile(connectionFile)
		if err != nil {
			return err
		}

		if encryptionKey != "" {
			re := regexp.MustCompile(`locations\/([a-zA-Z0-9_-]+)\/keyRings\/([a-zA-Z0-9_-]+)\/cryptoKeys\/([a-zA-Z0-9_-]+)`)
			ok := re.Match([]byte(encryptionKey))
			if !ok {
				return fmt.Errorf("encryption key must be of the format " +
					"locations/{location}/keyRings/{test}/cryptoKeys/{cryptoKey}")
			}
		}

		_, err = connections.Create(name, content, serviceAccountName,
			serviceAccountProject, encryptionKey, grantPermission, createSecret, wait)

		return
	},
}

var connectionFile, serviceAccountName, serviceAccountProject, encryptionKey string
var grantPermission, wait, createSecret bool

func init() {
	var name string
	CreateCmd.Flags().StringVarP(&name, "name", "n",
		"", "Connection name")
	CreateCmd.Flags().StringVarP(&connectionFile, "file", "f",
		"", "Connection details JSON file path")
	CreateCmd.Flags().BoolVarP(&grantPermission, "grant-permission", "g",
		false, "Grant the service account permission to the GCP resource; default is false")
	CreateCmd.Flags().StringVarP(&serviceAccountName, "sa", "",
		"", "Service Account name for the connection; do not include @<project-id>.iam.gserviceaccount.com")
	CreateCmd.Flags().StringVarP(&serviceAccountProject, "sp", "",
		"", "Service Account Project for the connection. Default is the connection's project id")
	CreateCmd.Flags().StringVarP(&encryptionKey, "encryption-keyid", "k",
		"", "Cloud KMS key for decrypting Auth Config; Format = locations/*/keyRings/*/cryptoKeys/*")
	CreateCmd.Flags().BoolVarP(&wait, "wait", "",
		false, "Waits for the connector to finish, with success or error; default is false")
	CreateCmd.Flags().BoolVarP(&createSecret, "create-secret", "",
		false, "Create Secret Manager secrets when creating the connection; default is false")

	_ = CreateCmd.MarkFlagRequired("name")
	_ = CreateCmd.MarkFlagRequired("file")
}
