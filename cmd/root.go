// Copyright 2020 Google LLC
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

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/srinandan/integrationcli/cmd/authconfigs"
	"github.com/srinandan/integrationcli/cmd/certificates"
	"github.com/srinandan/integrationcli/cmd/connectors"
	"github.com/srinandan/integrationcli/cmd/integrations"
	"github.com/srinandan/integrationcli/cmd/preferences"
	"github.com/srinandan/integrationcli/cmd/sfdcchannels"
	"github.com/srinandan/integrationcli/cmd/sfdcinstances"
	"github.com/srinandan/integrationcli/cmd/token"

	"github.com/srinandan/integrationcli/apiclient"

	"github.com/apigee/apigeecli/clilog"
	"github.com/spf13/cobra"
)

// RootCmd to manage apigeecli
var RootCmd = &cobra.Command{
	Use:   "integrationcli",
	Short: "Utility to work with Integration & Connectors",
	Long:  "This command lets you interact with Integration and Connector APIs.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		apiclient.SetServiceAccount(serviceAccount)
		apiclient.SetIntegrationToken(accessToken)

		if !disableCheck {
			if ok, _ := apiclient.TestAndUpdateLastCheck(); !ok {
				latestVersion, _ := getLatestVersion()
				if cmd.Version == "" {
					clilog.Info.Println("integrationcli wasn't built with a valid Version tag.")
				} else if latestVersion != "" && cmd.Version != latestVersion {
					fmt.Printf("You are using %s, the latest version %s is available for download\n", cmd.Version, latestVersion)
				}
			}
		}

		if useApigee {
			apiclient.UseApigeeIntegration()
		}

		_ = apiclient.SetAccessToken()

		return nil
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var accessToken, serviceAccount string
var disableCheck, useApigee, noOutput, verbose bool

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(&accessToken, "token", "t",
		"", "Google OAuth Token")

	RootCmd.PersistentFlags().StringVarP(&serviceAccount, "account", "a",
		"", "Path Service Account private key in JSON")

	RootCmd.PersistentFlags().BoolVarP(&disableCheck, "disable-check", "",
		false, "Disable check for newer versions")

	RootCmd.PersistentFlags().BoolVarP(&useApigee, "apigee-integration", "",
		false, "Use Apigee Integration; default is false (Application Integration)")

	RootCmd.PersistentFlags().BoolVarP(&noOutput, "no-output", "",
		false, "Disable printing API responses from the control plane")

	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "",
		false, "Enable verbose output from integrationcli")

	RootCmd.AddCommand(integrations.Cmd)
	RootCmd.AddCommand(preferences.Cmd)
	RootCmd.AddCommand(authconfigs.Cmd)
	RootCmd.AddCommand(connectors.Cmd)
	RootCmd.AddCommand(token.Cmd)
	RootCmd.AddCommand(certificates.Cmd)
	RootCmd.AddCommand(sfdcinstances.Cmd)
	RootCmd.AddCommand(sfdcchannels.Cmd)
}

func initConfig() {
	var skipLogInfo = true
	var skipCache bool

	if os.Getenv("INTEGRATIONCLI_SKIPLOG") == "false" || verbose {
		skipLogInfo = false
	}

	skipCache, _ = strconv.ParseBool(os.Getenv("INTEGRATIONCLI_SKIPCACHE"))

	apiclient.NewIntegrationClient(apiclient.IntegrationClientOptions{
		SkipCheck:   true,
		PrintOutput: true,
		SkipLogInfo: skipLogInfo,
		SkipCache:   skipCache,
	})
}

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd() *cobra.Command {
	return RootCmd
}

func getLatestVersion() (version string, err error) {
	var req *http.Request
	const endpoint = "https://api.github.com/repos/srinandan/integrationcli/releases/latest"

	client := &http.Client{}
	contentType := "application/json"

	req, err = http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", err
	}

	if result["tag_name"] == "" {
		clilog.Info.Println("Unable to determine latest tag, skipping this information")
		return "", nil
	} else {
		return fmt.Sprintf("%s", result["tag_name"]), nil
	}
}
