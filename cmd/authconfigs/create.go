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
	"os"
	"path"
	"regexp"

	"internal/apiclient"
	"internal/cloudkms"

	"internal/client/authconfigs"

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

			content, err = os.ReadFile(authConfigFile)
			if err != nil {
				return err
			}
		} else {

			if encryptionKey != "" {
				re := regexp.MustCompile(`locations\/([a-zA-Z0-9_-]+)\/keyRings\/([a-zA-Z0-9_-]+)\/cryptoKeys\/([a-zA-Z0-9_-]+)`)
				ok := re.Match([]byte(encryptionKey))
				if !ok {
					return fmt.Errorf("encryption key must be of the format locations/{location}/keyRings/{test}/cryptoKeys/{cryptoKey}")
				}
			}

			if _, err := os.Stat(encryptedFile); err != nil {
				return err
			}

			encryptedContent, err := os.ReadFile(encryptedFile)
			if err != nil {
				return err
			}

			fullEncryptionKey := path.Join("projects", apiclient.GetProjectID(), encryptionKey)
			content, err = cloudkms.DecryptSymmetric(fullEncryptionKey, encryptedContent)
			if err != nil {
				return err
			}
		}

		_, err = authconfigs.Create(content)
		return
	},
}

var authConfigFile, encryptedFile, encryptionKey string

func init() {
	CreateCmd.Flags().StringVarP(&authConfigFile, "file", "f",
		"", "Auth Config JSON file path")
	CreateCmd.Flags().StringVarP(&encryptedFile, "encrypted-file", "e",
		"", "Base64 encoded, Cloud KMS encrypted Auth Config JSON file path")
	CreateCmd.Flags().StringVarP(&encryptionKey, "encryption-keyid", "k",
		"", "Cloud KMS key for decrypting Auth Config; Format = locations/*keyRings/*/cryptoKeys/*")
}
