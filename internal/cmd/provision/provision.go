// Copyright 2023 Google LLC
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

package provision

import (
	"fmt"
	"internal/apiclient"
	"internal/client/provision"
	"internal/cmd/utils"
	"regexp"

	"github.com/spf13/cobra"
)

// Cmd to provision App Integration
var Cmd = &cobra.Command{
	Use:   "provision",
	Short: "Provisions application integration",
	Long:  "Provisions application integration in the region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")
		if err = apiclient.SetRegion(utils.GetStringParam(cmdRegion)); err != nil {
			return err
		}
		return apiclient.SetProjectID(utils.GetStringParam(cmdProject))
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cloudKMS := utils.GetStringParam(cmd.Flag("cloudkms"))
		serviceAccount := utils.GetStringParam(cmd.Flag("service-account"))

		if cloudKMS != "" {
			re := regexp.MustCompile(`projects\/([a-zA-Z0-9_-]+)\/locations\/([a-zA-Z0-9_-]+)\/` +
				`keyRings\/([a-zA-Z0-9_-]+)\/cryptoKeys\/([a-zA-Z0-9_-]+)\/cryptoKeyVersions\/([0-9]+)`)
			ok := re.Match([]byte(cloudKMS))
			if !ok {
				return fmt.Errorf("CloudKMS key must be of the format " +
					"projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{cryptoKey}" +
					"/cryptoKeyVersions/{cryptoKeyVersion}")
			}
		}

		if serviceAccount != "" {
			re := regexp.MustCompile(`[a-zA-Z0-9-]+@[a-zA-Z0-9-]+\.iam\.gserviceaccount\.com`)
			ok := re.Match([]byte(serviceAccount))
			if !ok {
				return fmt.Errorf("service account must of the format " +
					"<name>@<project-id>.iam.gserviceaccount.com")
			}
		}

		_, err = provision.Provision(cloudKMS, samples, gmek, serviceAccount)
		return err
	},
}

var samples, gmek bool

func init() {
	var cloudKMS, serviceAccount, project, region string

	Cmd.PersistentFlags().StringVarP(&project, "proj", "p",
		"", "Integration GCP Project name")
	Cmd.PersistentFlags().StringVarP(&region, "reg", "r",
		"", "Integration region name")
	Cmd.Flags().StringVarP(&cloudKMS, "cloudkms", "k",
		"", "Cloud KMS config for AuthModule to encrypt/decrypt credentials")
	Cmd.Flags().BoolVarP(&samples, "samples", "s",
		true, "Indicates if sample workflow should be created along with provisioning")
	Cmd.Flags().BoolVarP(&gmek, "gmek", "g",
		true, "Indicates provision with GMEK or CMEK")
	Cmd.Flags().StringVarP(&serviceAccount, "service-account", "",
		"", "User input run-as service account")
}
