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
	"internal/apiclient"
	"internal/client/connections"
	"os"
	"regexp"
	"strconv"

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
		createSecret, _ := strconv.ParseBool(cmd.Flag("create-secret").Value.String())
		grantPermission, _ := strconv.ParseBool(cmd.Flag("grant-permission").Value.String())
		wait, _ := strconv.ParseBool(cmd.Flag("wait").Value.String())
		name := cmd.Flag("name").Value.String()

		if _, err = os.Stat(connectionFile); err != nil {
			return fmt.Errorf("unable to open file %w", err)
		}

		content, err := os.ReadFile(connectionFile)
		if err != nil {
			return fmt.Errorf("unable to open file %w", err)
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

		return err
	},
	Example: `Create a PubSub connector and grant the Service Account permissions: ` + GetExample(0) + `
Create a GCS Connector: ` + GetExample(1),
}

var connectionFile, serviceAccountName, serviceAccountProject, encryptionKey string

func init() {
	var name string
	grantPermission, wait, createSecret := false, false, false

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
