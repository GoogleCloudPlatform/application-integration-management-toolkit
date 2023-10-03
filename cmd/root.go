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
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/endpoints"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/integrations"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/preferences"
	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd/provision"
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
		cmdServiceAccount := cmd.Flag("account").Value.String()
		cmdToken := cmd.Flag("token").Value.String()

		if metadataToken && (cmdServiceAccount != "" || cmdToken != "") {
			return fmt.Errorf("metadata-token cannot be used with token or account flags")
		}

		if cmdServiceAccount != "" && cmdToken != "" {
			return fmt.Errorf("token and account flags cannot be used together")
		}

		if !metadataToken {
			apiclient.SetServiceAccount(cmdServiceAccount)
			apiclient.SetIntegrationToken(cmdToken)
		}

		if !disableCheck {
			if ok, _ := apiclient.TestAndUpdateLastCheck(); !ok {
				latestVersion, _ := getLatestVersion()
				if cmd.Version == "" {
					clilog.Debug.Println("integrationcli wasn't built with a valid Version tag.")
				} else if latestVersion != "" && cmd.Version != latestVersion {
					clilog.Info.Printf("You are using %s, the latest version %s is available for download\n",
						cmd.Version, latestVersion)
				}
			}
		}

		apiclient.SetAPI(api)

		if metadataToken {
			return apiclient.GetDefaultAccessToken()
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

var (
	disableCheck, printOutput, noOutput, suppressWarnings, verbose, metadataToken bool
	api                                                                           apiclient.API
)

const ENABLED = "true"

func init() {
	var accessToken, serviceAccount string

	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(&accessToken, "token", "t",
		"", "Google OAuth Token")

	RootCmd.PersistentFlags().StringVarP(&serviceAccount, "account", "a",
		"", "Path Service Account private key in JSON")

	RootCmd.PersistentFlags().BoolVarP(&disableCheck, "disable-check", "",
		false, "Disable check for newer versions")

	RootCmd.PersistentFlags().BoolVarP(&printOutput, "print-output", "",
		true, "Control printing of info log statements")

	RootCmd.PersistentFlags().BoolVarP(&noOutput, "no-output", "",
		false, "Disable printing all statements to stdout")

	RootCmd.Flags().BoolVarP(&suppressWarnings, "suppress-warnings", "",
		false, "Disable printing warning statements to stdout")

	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "",
		false, "Enable verbose output from integrationcli")

	RootCmd.PersistentFlags().BoolVarP(&metadataToken, "metadata-token", "",
		false, "Metadata OAuth2 access token")

	RootCmd.PersistentFlags().Var(&api, "api", "Sets the control plane API. Must be one of prod, "+
		"staging or autopush; default is prod")

	RootCmd.AddCommand(integrations.Cmd)
	RootCmd.AddCommand(preferences.Cmd)
	RootCmd.AddCommand(authconfigs.Cmd)
	RootCmd.AddCommand(connectors.Cmd)
	RootCmd.AddCommand(token.Cmd)
	RootCmd.AddCommand(certificates.Cmd)
	RootCmd.AddCommand(sfdcinstances.Cmd)
	RootCmd.AddCommand(sfdcchannels.Cmd)
	RootCmd.AddCommand(endpoints.Cmd)
	RootCmd.AddCommand(provision.Cmd)
}

func initConfig() {
	debug := false
	var skipCache bool

	if os.Getenv("INTEGRATIONCLI_DEBUG") == ENABLED || verbose {
		debug = true
	}

	skipCache, _ = strconv.ParseBool(os.Getenv("INTEGRATIONCLI_SKIPCACHE"))

	if noOutput {
		printOutput = noOutput
	}

	if os.Getenv("INTEGRATIONCLI_DISABLE_RATELIMIT") == ENABLED {
		clilog.Debug.Println("integrationcli ratelimit is disabled")
		apiclient.SetRate(apiclient.None)
	} else {
		apiclient.SetRate(apiclient.IntegrationAPI)
	}

	apiclient.NewIntegrationClient(apiclient.IntegrationClientOptions{
		TokenCheck:    true,
		PrintOutput:   printOutput,
		NoOutput:      noOutput,
		DebugLog:      debug,
		SkipCache:     skipCache,
		MetadataToken: metadataToken,
	})
}

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd() *cobra.Command {
	return RootCmd
}

func getLatestVersion() (version string, err error) {
	var req *http.Request
	const endpoint = "https://api.github.com/repos/GoogleCloudPlatform/" +
		"application-integration-management-toolkit/releases/latest"

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
	return os.Getenv("INTEGRATIONCLI_NO_USAGE") == ENABLED
}

// getErrorsFlag
func getErrorsFlag() bool {
	return os.Getenv("INTEGRATIONCLI_NO_ERRORS") == ENABLED
}
