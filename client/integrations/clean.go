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
	"encoding/json"
	"fmt"

	"github.com/apigee/apigeecli/clilog"
	"github.com/srinandan/integrationcli/apiclient"
)

func Clean(name string, reportOnly bool, keepList []string) (err error) {

	var listOfVersions []basicIntegrationVersion
	var nextPage string

	apiclient.SetPrintOutput(false)
	for {
		respBody, err := ListVersions(name, -1, nextPage, "", "", false, false, true)
		if err != nil {
			return err
		}
		iversions := listbasicIntegrationVersions{}
		err = json.Unmarshal(respBody, &iversions)
		if err != nil {
			return err
		}

		listOfVersions = append(listOfVersions, iversions.BasicIntegrationVersions...)
		if iversions.NextPageToken == "" {
			break
		}
		nextPage = iversions.NextPageToken
	}

	if len(listOfVersions) == 0 {
		clilog.Warning.Println("no integration versions where found")
		return nil
	}

	for _, iversion := range listOfVersions {
		if iversion.State != "ACTIVE" {
			if reportOnly {
				fmt.Println("[REPORT]: Integration '" + name + "' Version: " + iversion.Version + " and Snapshot " + iversion.SnapshotNumber + " can be cleaned")
			} else {
				_, err = Delete(name, iversion.Version)
				if err != nil {
					return err
				}
			}
		}
	}

	apiclient.SetPrintOutput(true)

	return nil
}
