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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

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

		if encryptionKey != "" {
			re := regexp.MustCompile(`locations\/([a-zA-Z0-9_-]+)\/keyRings\/([a-zA-Z0-9_-]+)\/cryptoKeys\/([a-zA-Z0-9_-]+)`)
			ok := re.Match([]byte(encryptionKey))
			if !ok {
				return fmt.Errorf("encryption key must be of the format locations/{location}/keyRings/{test}/cryptoKeys/{cryptoKey}")
			}
		}

		type status struct {
			Code    int    `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		}

		type operation struct {
			Name     string  `json:"name,omitempty"`
			Done     bool    `json:"done,omitempty"`
			Error    *status `json:"error,omitempty"`
			Response *string `json:"response,omitempty"`
		}

		operationsBytes, err := connections.Create(name, content, serviceAccountName, serviceAccountProject, encryptionKey, grantPermission)

		if wait {
			o := operation{}
			if err = json.Unmarshal(operationsBytes, &o); err != nil {
				return err
			}

			fmt.Printf("Checking connector status in %d seconds\n", interval)

			apiclient.SetPrintOutput(false)

			stop := apiclient.Every(interval*time.Second, func(time.Time) bool {
				var respBody []byte

				if respBody, err = connections.GetOperation(o.Name); err != nil {
					return false
				}

				if err = json.Unmarshal(respBody, &o); err != nil {
					return false
				}

				if o.Done {
					if o.Error != nil {
						fmt.Printf("Connection completed with error: %v\n", o.Error)
					} else {
						fmt.Println("Connection completed successfully!")
					}
					return false
				} else {
					fmt.Printf("Connection status is: %t. Waiting %d seconds.\n", o.Done, interval)
					return true
				}
			})

			<-stop
		}
		return
	},
}

var connectionFile, serviceAccountName, serviceAccountProject, encryptionKey string
var grantPermission, wait bool

const interval = 10

func init() {
	CreateCmd.Flags().StringVarP(&name, "name", "n",
		"", "Connection name")
	CreateCmd.Flags().StringVarP(&connectionFile, "file", "f",
		"", "Connection details JSON file path")
	CreateCmd.Flags().BoolVarP(&grantPermission, "grant-permission", "g",
		false, "Grant the service account permission to the GCP resource")
	CreateCmd.Flags().StringVarP(&serviceAccountName, "sa", "",
		"", "Service Account name for the connection")
	CreateCmd.Flags().StringVarP(&serviceAccountProject, "sp", "",
		"", "Service Account Project for the connection. Default is the connection's project id")
	CreateCmd.Flags().StringVarP(&encryptionKey, "encryption-keyid", "k",
		"", "Cloud KMS key for decrypting Auth Config; Format = locations/*/keyRings/*/cryptoKeys/*")
	CreateCmd.Flags().BoolVarP(&wait, "wait", "",
		false, "Waits for the connector to finish, with success or error")

	_ = CreateCmd.MarkFlagRequired("name")
	_ = CreateCmd.MarkFlagRequired("file")
}
