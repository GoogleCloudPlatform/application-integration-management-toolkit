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

package apiclient

import (
	"fmt"
	"os"

	"github.com/apigee/apigeecli/clilog"
)

// BaseURL is the Integration control plane endpoint
const integrationBaseURL = "https://%s-integrations.googleapis.com/v1/projects/%s/locations/%s/products/apigee/"
const appIntegrationBaseURL = "https://%s-integrations.googleapis.com/v1/projects/%s/locations/%s/"
const connectorBaseURL = "https://connectors.googleapis.com/v1/projects/%s/locations/%s/connections"
const connectorOperationsBaseURL = "https://connectors.googleapis.com/v1/projects/%s/locations/%s/operations"

var integrationRegions = []string{"us", "us-west1"}

// IntegrationClientOptions is the base struct to hold all command arguments
type IntegrationClientOptions struct {
	Region               string //Integration region
	Token                string //Google OAuth access token
	ServiceAccount       string //Google service account json
	ProjectID            string //GCP Project ID
	SkipLogInfo          bool   //LogInfo controls the log level
	SkipCheck            bool   //skip checking access token expiry
	SkipCache            bool   //skip writing access token to file
	PrintOutput          bool   //prints output from http calls
	ProxyUrl             string //use a proxy url
	ExportToFile         string //determine of the contents should be written to file
	UseApigeeIntegration bool   //use Apigee Integration; defaults to Application Integration
	NoOutput             bool   //disables printing API responses
}

var options *IntegrationClientOptions

// NewIntegrationClient sets up options to invoke Integration APIs
func NewIntegrationClient(o IntegrationClientOptions) {
	if options == nil {
		options = new(IntegrationClientOptions)
	}

	if o.Region != "" {
		options.Region = o.Region
	}
	if o.Token != "" {
		options.Token = o.Token
	}
	if o.ServiceAccount != "" {
		options.ServiceAccount = o.ServiceAccount
	}
	if o.ProjectID != "" {
		options.ProjectID = o.ProjectID
	}
	if o.SkipCheck {
		options.SkipCheck = true
	} else {
		options.SkipCheck = false
	}
	if o.SkipCache {
		options.SkipCache = true
	} else {
		options.SkipCache = false
	}
	if o.SkipLogInfo {
		options.SkipLogInfo = true
		clilog.Init(true)
	} else {
		options.SkipLogInfo = false
		clilog.Init(false)
	}
	if o.PrintOutput {
		options.PrintOutput = true
	} else {
		options.PrintOutput = false
	}
	if o.ExportToFile != "" {
		options.ExportToFile = o.ExportToFile
	}
	if o.NoOutput {
		options.NoOutput = true
	} else {
		options.NoOutput = false
	}
	//read preference file
	_ = ReadPreferencesFile()

}

// SetRegion sets the org variable
func SetRegion(region string) (err error) {
	if region == "" {
		if GetRegion() == "" {
			return fmt.Errorf("region was not set in preferences or supplied in the command")
		}
		return nil
	}
	options.Region = region
	return nil
}

// GetRegion gets the org variable
func GetRegion() string {
	return options.Region
}

// SetIntegrationToken sets the access token for use with Integration API calls
func SetIntegrationToken(token string) {
	options.Token = token
}

// GetIntegrationToken get the access token value in client opts (does not generate it)
func GetIntegrationToken() string {
	return options.Token
}

// SetProjectID sets the project id
func SetProjectID(projectID string) (err error) {
	if projectID == "" {
		if GetProjectID() == "" {
			return fmt.Errorf("projectId was not set in preferences or supplied in the command")
		}
		return nil
	}
	options.ProjectID = projectID
	return nil
}

// GetProjectID gets the project id
func GetProjectID() string {
	return options.ProjectID
}

// SetServiceAccount
func SetServiceAccount(serviceAccount string) {
	options.ServiceAccount = serviceAccount
}

// GetServiceAccount
func GetServiceAccount() string {
	if options.ServiceAccount == "" {
		envVar := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if envVar != "" {
			options.ServiceAccount = envVar
		}
	}
	return options.ServiceAccount
}

// IsSkipCheck
func IsSkipCheck() bool {
	return options.SkipCheck
}

// IsSkipCache
func IsSkipCache() bool {
	return options.SkipCache
}

// IsSkipLogInfo
func IsSkipLogInfo() bool {
	return options.SkipLogInfo
}

// SetSkipLogIngo
func SetSkipLogInfo(l bool) {
	options.SkipLogInfo = l
	clilog.Init(l)
}

// PrintOutput
func SetPrintOutput(output bool) {
	options.PrintOutput = output
}

// GetPrintOutput
func GetPrintOutput() bool {
	return options.PrintOutput
}

// GetProxyURL
func GetProxyURL() string {
	return options.ProxyUrl
}

// SetProxyURL
func SetProxyURL(proxyurl string) {
	options.ProxyUrl = proxyurl
}

// GetBaseIntegrationURL
func GetBaseIntegrationURL() (integrationUrl string) {
	if options.ProjectID == "" || options.Region == "" {
		return ""
	}
	if options.UseApigeeIntegration {
		return fmt.Sprintf(integrationBaseURL, GetRegion(), GetProjectID(), GetRegion())
	}
	return fmt.Sprintf(appIntegrationBaseURL, GetRegion(), GetProjectID(), GetRegion())
}

// GetBaseConnectorURL
func GetBaseConnectorURL() (connectorUrl string) {
	if options.ProjectID == "" || options.Region == "" {
		return ""
	}
	return fmt.Sprintf(connectorBaseURL, GetProjectID(), GetRegion())
}

// GetBaseConnectorOperationsURL
func GetBaseConnectorOperationsrURL() (connectorUrl string) {
	if options.ProjectID == "" || options.Region == "" {
		return ""
	}
	return fmt.Sprintf(connectorOperationsBaseURL, GetProjectID(), GetRegion())
}

// GetIntegrationRegions
func GetIntegrationRegions() []string {
	return integrationRegions
}

// SetExportToFile
func SetExportToFile(exportToFile string) {
	options.ExportToFile = exportToFile
}

// GetExportToFile
func GetExportToFile() string {
	return options.ExportToFile
}

// DryRun
func DryRun() bool {
	if os.Getenv("INTEGRATIONCLI_DRYNRUN") != "" {
		clilog.Warning.Println("Dry run mode enabled! unset INTEGRATIONCLI_DRYNRUN to disable dry run")
		return true
	}
	return false
}

func UseApigeeIntegration() {
	options.UseApigeeIntegration = true
}
