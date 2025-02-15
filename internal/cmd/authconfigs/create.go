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
	"errors"
	"fmt"
	"internal/apiclient"
	"internal/client/authconfigs"
	"internal/clilog"
	"internal/cloudkms"
	"internal/cmd/utils"
	"os"
	"path"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CreateCmd to create authconfigs
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an authconfig",
	Long:  "Create an authconfig",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		project := utils.GetStringParam(cmd.Flag("proj"))
		region := utils.GetStringParam(cmd.Flag("reg"))

		if err = apiclient.SetRegion(region); err != nil {
			return err
		}

		if authConfigFile != "" && (encryptedFile != "" || encryptionKey != "") {
			return errors.New("file cannot be combined with encrypted-file or encryption-keyid")
		}

		if (encryptedFile != "" && encryptionKey == "") || (encryptedFile == "" && encryptionKey != "") {
			return errors.New("encrypted-file and encryption-keyid must both be set")
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

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
					return fmt.Errorf("encryption key must be of the format " +
						"locations/{location}/keyRings/{test}/cryptoKeys/{cryptoKey}")
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

		return apiclient.PrettyPrint(authconfigs.Create(content))
	},
	Example: `Create a new user name auth config: ` + GetExample(0) + `
Create a new OIDC auth config: ` + GetExample(1) + `
Create a new auth token auth config: ` + GetExample(2) + `
Create a new auth config from Cloud KMS Encrypted files: ` + GetExample(3),
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
