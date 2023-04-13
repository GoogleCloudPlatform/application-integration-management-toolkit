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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/authconfigs"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/certificates"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/connectors"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/integrations"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/preferences"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/sfdcchannels"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/sfdcinstances"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/token"

	"internal/apiclient"

	"internal/clilog"

	"github.com/spf13/cobra"
)

// RootCmd to manage integrationcli
var RootCmd = &cobra.Command{
	Use:   "integrationcli",
	Short: "Utility to work with Integration & Connectors",
	Long:  "This command lets you interact with Integration and Connector APIs.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmdServiceAccount := cmd.Flag("account")
		cmdToken := cmd.Flag("token")

		apiclient.SetServiceAccount(cmdServiceAccount.Value.String())
		apiclient.SetIntegrationToken(cmdToken.Value.String())

		if !disableCheck {
			if ok, _ := apiclient.TestAndUpdateLastCheck(); !ok {
				latestVersion, _ := getLatestVersion()
				if cmd.Version == "" {
					clilog.Debug.Println("integrationcli wasn't built with a valid Version tag.")
				} else if latestVersion != "" && cmd.Version != latestVersion {
					clilog.Info.Printf("You are using %s, the latest version %s is available for download\n", cmd.Version, latestVersion)
				}
			}
		}

		if useApigee {
			apiclient.UseApigeeIntegration()
		}

		_ = apiclient.SetAccessToken()

		return nil
	},
	SilenceUsage:  getUsageFlag(),
	SilenceErrors: getErrorsFlag(),
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		clilog.Error.Println(err)
	}
}

var disableCheck, useApigee, printOutput, noOutput, verbose bool

func init() {
	var accessToken, serviceAccount string

	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(&accessToken, "token", "t",
		"", "Google OAuth Token")

	RootCmd.PersistentFlags().StringVarP(&serviceAccount, "account", "a",
		"", "Path Service Account private key in JSON")

	RootCmd.PersistentFlags().BoolVarP(&disableCheck, "disable-check", "",
		false, "Disable check for newer versions")

	RootCmd.PersistentFlags().BoolVarP(&useApigee, "apigee-integration", "",
		false, "Use Apigee Integration; default is false (Application Integration)")

	RootCmd.PersistentFlags().BoolVarP(&printOutput, "print-output", "",
		true, "Control printing of info log statements")

	RootCmd.PersistentFlags().BoolVarP(&noOutput, "no-output", "",
		false, "Disable printing all statements to stdout")

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
	var debug = false
	var skipCache bool

	if os.Getenv("INTEGRATIONECLI_DEBUG") == "true" || verbose {
		debug = true
	}

	skipCache, _ = strconv.ParseBool(os.Getenv("INTEGRATIONCLI_SKIPCACHE"))

	if noOutput {
		printOutput = noOutput
	}

	apiclient.NewIntegrationClient(apiclient.IntegrationClientOptions{
		TokenCheck:  true,
		PrintOutput: printOutput,
		NoOutput:    noOutput,
		DebugLog:    debug,
		SkipCache:   skipCache,
	})
}

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd() *cobra.Command {
	return RootCmd
}

func getLatestVersion() (version string, err error) {
	var req *http.Request
	const endpoint = "https://api.github.com/repos/GoogleCloudPlatform/application-integration-management-toolkit/releases/latest"

	client := &http.Client{}
	contentType := "application/json"

	ctx := context.Background()

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", err
	}

	if result["tag_name"] == "" {
		clilog.Debug.Println("Unable to determine latest tag, skipping this information")
		return "", nil
	}
	return fmt.Sprintf("%s", result["tag_name"]), nil

}

// getUsageFlag
func getUsageFlag() bool {
	return os.Getenv("INTEGRATIONCLI_NO_USAGE") == "true"
}

// getErrorsFlag
func getErrorsFlag() bool {
	return os.Getenv("INTEGRATIONCLI_NO_ERRORS") == "true"
}
