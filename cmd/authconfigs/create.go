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

package authconfigs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/srinandan/integrationcli/apiclient"
	"github.com/srinandan/integrationcli/client/authconfigs"
	"github.com/srinandan/integrationcli/cloudkms"

	"github.com/spf13/cobra"
)

// CreateCmd to create authconfigs
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an authconfig",
	Long:  "Create an authconfig",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}

		if authConfigFile != "" && (encryptedFile != "" || encryptionKey != "") {
			return fmt.Errorf("file cannot be combined with encrypted-file or encryption-keyid")
		}

		if (encryptedFile != "" && encryptionKey == "") || (encryptedFile == "" && encryptionKey != "") {
			return fmt.Errorf("encrypted-file and encryption-keyid must both be set")
		}

		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		var content []byte

		if authConfigFile != "" {
			if _, err := os.Stat(authConfigFile); err != nil {
				return err
			}

			content, err = ioutil.ReadFile(authConfigFile)
			if err != nil {
				return err
			}
		} else {
			if _, err := os.Stat(encryptedFile); err != nil {
				return err
			}

			encryptedContent, err := ioutil.ReadFile(encryptedFile)
			if err != nil {
				return err
			}

			fullEncryptionKey := path.Join("projects", apiclient.GetProjectID(), "locations", apiclient.GetRegion(), encryptionKey)
			content, err = cloudkms.DecryptSymmetric(fullEncryptionKey, encryptedContent)
			if err != nil {
				return err
			}
		}

		_, err = authconfigs.Create(name, content)
		return
	},
}

var authConfigFile, encryptedFile, encryptionKey string

func init() {
	CreateCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	CreateCmd.Flags().StringVarP(&authConfigFile, "file", "f",
		"", "Auth Config JSON file path")
	CreateCmd.Flags().StringVarP(&encryptedFile, "encrypted-file", "e",
		"", "Base64 encoded, Cloud KMS encrypted Auth Config JSON file path")
	CreateCmd.Flags().StringVarP(&encryptionKey, "encryption-keyid", "k",
		"", "Cloud KMS key for decrypting Auth Config; Format = keyRings/*/cryptoKeys/*")

	_ = CreateCmd.MarkFlagRequired("name")
}
