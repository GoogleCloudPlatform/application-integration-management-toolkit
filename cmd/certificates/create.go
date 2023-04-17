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

package certificates

import (
	"errors"
	"os"

	"internal/apiclient"

	"internal/client/certificates"

	"github.com/spf13/cobra"
)

// CreateCmd to create authconfigs
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a certificate entity in Application integration",
	Long:  "Create a certificate entity in Application integration",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		project := cmd.Flag("proj").Value.String()
		region := cmd.Flag("reg").Value.String()
		privateKeyFile := cmd.Flag("private-key").Value.String()
		passphrase := cmd.Flag("passphrase").Value.String()

		if err = apiclient.SetRegion(region); err != nil {
			return err
		}

		if passphrase != "" && privateKeyFile == "" {
			return errors.New("private key must be used with passphrase")
		}

		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var sslCertContent, privateKeyCertContent []byte
		name := cmd.Flag("name").Value.String()
		description := cmd.Flag("description").Value.String()
		sslCertificateFile := cmd.Flag("cert-file").Value.String()
		privateKeyFile := cmd.Flag("private-key").Value.String()
		passphrase := cmd.Flag("passphrase").Value.String()

		if sslCertificateFile != "" {
			if _, err := os.Stat(sslCertificateFile); err != nil {
				return err
			}

			sslCertContent, err = os.ReadFile(sslCertificateFile)
			if err != nil {
				return err
			}
		}

		if privateKeyFile != "" {
			if _, err := os.Stat(privateKeyFile); err != nil {
				return err
			}

			privateKeyCertContent, err = os.ReadFile(privateKeyFile)
			if err != nil {
				return err
			}
		}

		_, err = certificates.Create(name, description, string(sslCertContent), string(privateKeyCertContent), passphrase)
		return
	},
}

func init() {
	var name, description, sslCertificateFile, privateKeyFile, passphrase string

	CreateCmd.Flags().StringVarP(&name, "name", "n",
		"", "Display name for the certificate")
	CreateCmd.Flags().StringVarP(&description, "description", "d",
		"", "Description for the certificate")

	CreateCmd.Flags().StringVarP(&sslCertificateFile, "cert-file", "",
		"", "Path to TLS Certificate file (PEM) format")
	CreateCmd.Flags().StringVarP(&privateKeyFile, "private-key", "",
		"", "Path to TLS Private Key file (PEM) format")
	CreateCmd.Flags().StringVarP(&passphrase, "passphrase", "",
		"", "Passphrase for the private key")

	_ = CreateCmd.MarkFlagRequired("name")
	_ = CreateCmd.MarkFlagRequired("cert-file")
}
