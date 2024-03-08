// Copyright 2024 Google LLC
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

package clienttest

import (
	"fmt"
	"os"

	"internal/apiclient"
)

func TestSetup() (err error) {
	apiclient.NewIntegrationClient(apiclient.IntegrationClientOptions{
		TokenCheck:  true,
		PrintOutput: true,
		NoOutput:    false,
		DebugLog:    true,
		SkipCache:   true,
	})

	proj := os.Getenv("PROJECT_ID")
	if err = apiclient.SetProjectID(proj); err != nil {
		return fmt.Errorf("PROJECT_ID not set")
	}

	region := os.Getenv("LOCATION")
	if region != "" {
		apiclient.SetRegion(region)
	} else {
		return fmt.Errorf("LOCATION not set")
	}
	token := os.Getenv("INTEGRATION_TOKEN")
	if token == "" {
		err = apiclient.GetDefaultAccessToken()
		if err != nil {
			return fmt.Errorf("INTEGRATION_TOKEN not set")
		}
	} else {
		apiclient.SetIntegrationToken(token)
	}

	cliPath := os.Getenv("INTEGRATIONCLI_PATH")
	if cliPath == "" {
		return fmt.Errorf("INTEGRATIONCLI_PATH not set")
	}

	return nil
}
