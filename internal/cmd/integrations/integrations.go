// Copyright 2022 Google LLC
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

package integrations

import (
	"fmt"
	"internal/apiclient"
	"internal/client/integrations"
	"internal/clilog"
	"internal/cmd/utils"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// Cmd to manage integration flows
var Cmd = &cobra.Command{
	Use:   "integrations",
	Short: "Manage integrations in a GCP project",
	Long:  "Manage integrations in a GCP project",
}

// var userLabel, snapshot string
var examples = []string{
	`integrationcli integrations create -n $name -f samples/sample.json -u $userLabel --default-token`,
	`integrationcli integrations create -n $name -f samples/sample.json -o samples/sample_overrides.json --default-token`,
	`integrationcli integrations create -n $name -f samples/sample.json --publish=true --default-token`,
	`integrationcli integrations versions list -n $integration --basic=true --default-token`,
	`integrationcli integrations versions list -n $integration --basic=true --filter=state=ACTIVE --default-token`,
	`integrationcli integrations scaffold -n $integration -s $snapshot -f . --env=dev --default-token`,
	`integrationcli integrations scaffold -n $integration -s $snapshot -f . --skip-connectors=true --default-token`,
	`integrationcli integrations scaffold -n $integration -s $snapshot -f . --cloud-build=true --default-token`,
	`integrationcli integrations scaffold -n $integration -s $snapshot -f . --cloud-deploy=true --default-token`,
	`integrationcli integrations apply -f . --wait=true --default-token`,
	`integrationcli integrations apply -f . --env=dev --default-token`,
	`integrationcli integrations apply -f . --grant-permission=true --default-token`,
	`integrationcli integrations apply -f . --skip-connectors=true --default-token`,
	`integrationcli integrations create -n $name -f samples/sample.json --basic=true --default-token`,
	`integrationcli integrations versions publish -n $name --default-token`,
	`integrationcli integrations versions publish -n $name -s $snapshot --default-token`,
	`integrationcli integrations versions unpublish -n $name --default-token`,
	`integrationcli integrations versions unpublish -n $name -u $userLabel --default-token`,
	`integrationcli integrations apply -f . --env=dev --tests-folder=./test-configs --default-token`,
}

func init() {
	var project, region string

	Cmd.PersistentFlags().StringVarP(&project, "proj", "p",
		"", "Integration GCP Project name")

	Cmd.PersistentFlags().StringVarP(&region, "reg", "r",
		"", "Integration region name")

	Cmd.AddCommand(ListCmd)
	Cmd.AddCommand(VerCmd)
	Cmd.AddCommand(CleanCmd)
	Cmd.AddCommand(ExecuteCmd)
	Cmd.AddCommand(ExecCmd)
	Cmd.AddCommand(ExportCmd)
	Cmd.AddCommand(ImportCmd)
	Cmd.AddCommand(UploadCmd)
	Cmd.AddCommand(CreateCmd)
	Cmd.AddCommand(DelCmd)
	Cmd.AddCommand(ScaffoldCmd)
	Cmd.AddCommand(ApplyCmd)
	Cmd.AddCommand(TestCasesCmd)
}

func GetExample(i int) string {
	return examples[i]
}

func getLatestVersion(name string) (version string, err error) {
	var listBody []byte

	apiclient.DisableCmdPrintHttpResponse()
	defer apiclient.EnableCmdPrintHttpResponse()

	// list integration versions, order by state=ACTIVE, page size = 1 and return basic info
	if listBody, err = integrations.ListVersions(name, 1, "", "state=ACTIVE",
		"snapshot_number", false, false, true); err != nil {
		return "", fmt.Errorf("unable to list versions: %v", err)
	}
	if string(listBody) != "{}" {
		if version, err = integrations.GetIntegrationVersion(listBody); err != nil {
			return "", err
		}
	} else {
		// list integration versions, order by state=SNAPSHOT, page size = 1 and return basic info
		if listBody, err = integrations.ListVersions(name, 1, "", "state=SNAPSHOT",
			"snapshot_number", false, false, true); err != nil {
			return "", fmt.Errorf("unable to list versions: %v", err)
		}
		if string(listBody) != "{}" {
			if version, err = integrations.GetIntegrationVersion(listBody); err != nil {
				return "", err
			}
		} else {
			if listBody, err = integrations.ListVersions(name, 1, "", "state=DRAFT",
				"snapshot_number", false, false, true); err != nil {
				return "", fmt.Errorf("unable to list versions: %v", err)
			}
			if string(listBody) != "{}" {
				if version, err = integrations.GetIntegrationVersion(listBody); err != nil {
					return "", err
				}
			}
		}
	}
	return version, nil
}

func executeAllTestCases(inputFolder string, name string, version string) (err error) {

	if stat, err := os.Stat(inputFolder); stat == nil || (err != nil && !stat.IsDir()) {
		return fmt.Errorf("supplied path is not a folder: %v", err)
	}

	rJSONFiles := regexp.MustCompile(`(\S*)\.json`)
	var inputFiles []string

	_ = filepath.Walk(inputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			inputFileName := filepath.Base(path)
			if rJSONFiles.MatchString(inputFileName) {
				clilog.Info.Printf("Found test case file %s for integration: %s\n", inputFileName, name)
				inputFiles = append(inputFiles, inputFileName)
			}
		}
		return nil
	})

	if len(inputFiles) > 0 {
		for _, inputFileName := range inputFiles {
			content, err := utils.ReadFile(path.Join(inputFolder, inputFileName))
			if err != nil {
				return err
			}
			testDisplayName := strings.TrimSuffix(filepath.Base(inputFileName), filepath.Ext(filepath.Base(inputFileName)))
			apiclient.ClientPrintHttpResponse.Set(false)
			testCaseID, err := integrations.FindTestCase(name, version, testDisplayName, "")
			apiclient.ClientPrintHttpResponse.Set(true)
			if err != nil {
				return err
			}
			clilog.Info.Printf("Executing test cases from file %s for integration: %s\n", inputFileName, name)
			testCaseResp, err := integrations.ExecuteTestCase(name, version, testCaseID, string(content))
			if err != nil {
				return err
			}
			err = integrations.AssertTestExecutionResult(testCaseResp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
