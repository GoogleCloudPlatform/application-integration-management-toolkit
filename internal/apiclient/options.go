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
	"strings"
	"sync"

	"internal/clilog"
)

// BaseURL is the Integration control plane endpoint
const (
	integrationBaseURL            = "https://%s-integrations.googleapis.com/v1/projects/%s/locations/%s/products/apigee/"
	appIntegrationAutoPushBaseURL = "https://autopushqual%s-integrations.sandbox.googleapis.com/v1/projects/%s/locations/%s/"
	appIntegrationStagingBaseURL  = "https://stagingqual%s-integrations.sandbox.googleapis.com/v1/projects/%s/locations/%s/"
	appIntegrationBaseURL         = "https://%s-integrations.googleapis.com/v1/projects/%s/locations/%s/"

	connectorBaseURL         = "https://connectors.googleapis.com/v1/projects/%s/locations/%s/connections"
	connectorAutoPushBaseURL = "https://autopush-connectors.sandbox.googleapis.com/v1/projects/%s/locations/%s/connections"
	connectorStagingBaseURL  = "https://staging-connectors.sandbox.googleapis.com/v1/projects/%s/locations/%s/connections"

	customConnectorBaseURL         = "https://connectors.googleapis.com/v1/projects/%s/locations/global/customConnectors"
	customConnectorAutoPushBaseURL = "https://autopush-connectors.sandbox.googleapis.com/v1/projects/%s/locations/global/customConnectors"
	customConnectorStagingBaseURL  = "https://staging-connectors.sandbox.googleapis.com/v1/projects/%s/locations/global/customConnectors"

	connectorOperationsBaseURL         = "https://connectors.googleapis.com/v1/projects/%s/locations/%s/operations"
	connectorOperationsAutoPushBaseURL = "https://autopush-connectors.sandbox.googleapis.com/v1/projects/%s/locations/%s/operations"
	connectorOperationsStagingBaseURL  = "https://staging-connectors.sandbox.googleapis.com/v1/projects/%s/locations/%s/operations"

	connectorEndpointAttachURL         = "https://connectors.googleapis.com/v1/projects/%s/locations/%s/endpointAttachments"
	connectorEndpointAttachAutoPushURL = "https://autopush-connectors.sandbox.googleapis.com/v1/projects/%s/locations/%s/endpointAttachments"
	connectorEndpointAttachStagingURL  = "https://staging-connectors.sandbox.googleapis.com/v1/projects/%s/locations/%s/endpointAttachments"

	connectorZonesURL         = "https://connectors.googleapis.com/v1/projects/%s/locations/global/managedZones"
	connectorZonesAutoPushURL = "https://autopush-connectors.sandbox.googleapis.com/v1/projects/%s/locations/global/managedZones"
	connectorZonesStagingURL  = "https://staging-connectors.sandbox.googleapis.com/v1/projects/%s/locations/global/managedZones"
)

// IntegrationClientOptions is the base struct to hold all command arguments
type IntegrationClientOptions struct {
	Api                API    // integrationcli can switch between prod, autopush and staging
	Region             string // Integration region
	Token              string // Google OAuth access token
	ServiceAccount     string // Google service account json
	ProjectID          string // GCP Project ID
	DebugLog           bool   // Enable debug logs
	TokenCheck         bool   // skip checking access token expiry
	SkipCache          bool   // skip writing access token to file
	PrintOutput        bool   // prints output from http calls
	NoOutput           bool   // Disable all statements to stdout
	SuppressWarnings   bool   // Disable printing of warnings to stdout
	ProxyUrl           string // use a proxy url
	MetadataToken      bool   // use metadata outh2 token
	ExportToFile       string // determine of the contents should be written to file
	ConflictsAreErrors bool   // treat statusconflict as an error
}

var options *IntegrationClientOptions

type Rate uint8

const (
	None Rate = iota
	IntegrationAPI
	ConnectorsAPI
)

type API string

const (
	PROD     API = "prod"
	STAGING  API = "staging"
	AUTOPUSH API = "autopush"
)

var apiRate Rate

var cmdPrintHttpResponses = true

type clientPrintHttpResponse struct {
	enable bool
	sync.Mutex
}

var ClientPrintHttpResponse = &clientPrintHttpResponse{enable: true}

// NewIntegrationClient sets up options to invoke Integration APIs
func NewIntegrationClient(o IntegrationClientOptions) {
	if options == nil {
		options = new(IntegrationClientOptions)
	}

	options.TokenCheck = o.TokenCheck
	options.SkipCache = o.SkipCache
	options.DebugLog = o.DebugLog
	options.PrintOutput = o.PrintOutput
	options.NoOutput = o.NoOutput
	options.SuppressWarnings = o.SuppressWarnings

	// initialize logs
	clilog.Init(options.DebugLog, options.PrintOutput, options.NoOutput, options.SuppressWarnings)

	cliPref, err := readPreferencesFile()
	if err != nil {
		clilog.Debug.Println(err)
	}

	if cliPref != nil {
		options.ProjectID = cliPref.Project
		options.Region = cliPref.Region
		options.ProxyUrl = cliPref.ProxyUrl
		options.Token = cliPref.Token
		options.TokenCheck = cliPref.Nocheck
		if cliPref.Api != "" {
			options.Api = cliPref.Api
		}
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
	if o.ExportToFile != "" {
		options.ExportToFile = o.ExportToFile
	}
	if o.Api == "" {
		options.Api = o.Api
	}

	options.ConflictsAreErrors = true
}

func (a *API) String() string {
	return string(*a)
}

func (a *API) Set(r string) error {
	switch r {
	case "prod", "staging", "autopush":
		*a = API(r)
	default:
		return fmt.Errorf("must be one of %s,%s or %s", PROD, STAGING, AUTOPUSH)
	}
	return nil
}

func (a *API) Type() string {
	return "api"
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

// TokenCheckEnabled
func TokenCheckEnabled() bool {
	return options.TokenCheck
}

// IsSkipCache
func IsSkipCache() bool {
	return options.SkipCache
}

// DebugEnabled
func DebugEnabled() bool {
	return options.DebugLog
}

// PrintOutput
func SetPrintOutput(output bool) {
	options.PrintOutput = output
}

// GetPrintOutput
func GetPrintOutput() bool {
	return options.PrintOutput
}

// DisableCmdPrintHttpResponse
func DisableCmdPrintHttpResponse() {
	cmdPrintHttpResponses = false
}

// EnableCmdPrintHttpResponse
func EnableCmdPrintHttpResponse() {
	cmdPrintHttpResponses = true
}

// GetPrintHttpResponseSetting
func GetCmdPrintHttpResponseSetting() bool {
	return cmdPrintHttpResponses
}

// SetClientPrintHttpResponse
func (c *clientPrintHttpResponse) Set(b bool) {
	c.Lock()
	defer c.Unlock()
	c.enable = b
}

// GetPrintHttpResponseSetting
func (c *clientPrintHttpResponse) Get() bool {
	c.Lock()
	defer c.Unlock()
	return c.enable
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
	switch options.Api {
	case PROD:
		return fmt.Sprintf(appIntegrationBaseURL, GetRegion(), GetProjectID(), GetRegion())
	case STAGING:
		// the url for staging is like:
		// https://stagingqualuswest1-integrations.sandbox.googleapis.com/v1/projects/-/locations/us-west1/integrations
		return fmt.Sprintf(appIntegrationStagingBaseURL, strings.Replace(GetRegion(), "-", "", -1), GetProjectID(), GetRegion())
	case AUTOPUSH:
		// the url for autopush is like:
		// https://autopushqualuswest1-integrations.sandbox.googleapis.com/v1/projects/-/locations/us-west1/integrations
		return fmt.Sprintf(appIntegrationAutoPushBaseURL, strings.Replace(GetRegion(), "-", "", -1), GetProjectID(), GetRegion())
	default:
		return fmt.Sprintf(appIntegrationBaseURL, GetRegion(), GetProjectID(), GetRegion())
	}
}

// GetBaseConnectorURL
func GetBaseConnectorURL() (connectorUrl string) {
	if options.ProjectID == "" || options.Region == "" {
		return ""
	}
	switch options.Api {
	case PROD:
		return fmt.Sprintf(connectorBaseURL, GetProjectID(), GetRegion())
	case STAGING:
		return fmt.Sprintf(connectorStagingBaseURL, GetProjectID(), GetRegion())
	case AUTOPUSH:
		return fmt.Sprintf(connectorAutoPushBaseURL, GetProjectID(), GetRegion())
	default:
		return fmt.Sprintf(connectorBaseURL, GetProjectID(), GetRegion())
	}
}

// GetBaseCustomConnectorURL
func GetBaseCustomConnectorURL() (connectorUrl string) {
	if options.ProjectID == "" || options.Region == "" {
		return ""
	}
	switch options.Api {
	case PROD:
		return fmt.Sprintf(customConnectorBaseURL, GetProjectID())
	case STAGING:
		return fmt.Sprintf(customConnectorStagingBaseURL, GetProjectID())
	case AUTOPUSH:
		return fmt.Sprintf(customConnectorAutoPushBaseURL, GetProjectID())
	default:
		return fmt.Sprintf(customConnectorBaseURL, GetProjectID())
	}
}

// GetBaseConnectorURLWithRegion
func GetBaseConnectorURLWithRegion(region string) (connectorUrl string) {
	if options.ProjectID == "" || region == "" {
		return ""
	}
	switch options.Api {
	case PROD:
		return fmt.Sprintf(connectorBaseURL, GetProjectID(), region)
	case STAGING:
		return fmt.Sprintf(connectorStagingBaseURL, GetProjectID(), region)
	case AUTOPUSH:
		return fmt.Sprintf(connectorAutoPushBaseURL, GetProjectID(), region)
	default:
		return fmt.Sprintf(connectorBaseURL, GetProjectID(), region)
	}
}

// GetBaseConnectorOperationsURL
func GetBaseConnectorOperationsrURL() (connectorUrl string) {
	if options.ProjectID == "" || options.Region == "" {
		return ""
	}
	switch options.Api {
	case PROD:
		return fmt.Sprintf(connectorOperationsBaseURL, GetProjectID(), GetRegion())
	case STAGING:
		return fmt.Sprintf(connectorOperationsStagingBaseURL, GetProjectID(), GetRegion())
	case AUTOPUSH:
		return fmt.Sprintf(connectorOperationsAutoPushBaseURL, GetProjectID(), GetRegion())
	default:
		return fmt.Sprintf(connectorOperationsBaseURL, GetProjectID(), GetRegion())
	}
}

// GetBaseConnectorEndpointAttachURL
func GetBaseConnectorEndpointAttachURL() (connectorUrl string) {
	if options.ProjectID == "" || options.Region == "" {
		return ""
	}
	switch options.Api {
	case PROD:
		return fmt.Sprintf(connectorEndpointAttachURL, GetProjectID(), GetRegion())
	case STAGING:
		return fmt.Sprintf(connectorEndpointAttachStagingURL, GetProjectID(), GetRegion())
	case AUTOPUSH:
		return fmt.Sprintf(connectorEndpointAttachAutoPushURL, GetProjectID(), GetRegion())
	default:
		return fmt.Sprintf(connectorEndpointAttachURL, GetProjectID(), GetRegion())
	}
}

// GetBaseConnectorZonesURL
func GetBaseConnectorZonesURL() (connectorUrl string) {
	if options.ProjectID == "" || options.Region == "" {
		return ""
	}
	switch options.Api {
	case PROD:
		return fmt.Sprintf(connectorZonesURL, GetProjectID())
	case STAGING:
		return fmt.Sprintf(connectorZonesAutoPushURL, GetProjectID())
	case AUTOPUSH:
		return fmt.Sprintf(connectorZonesAutoPushURL, GetProjectID())
	default:
		return fmt.Sprintf(connectorZonesURL, GetProjectID())
	}
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

// SetNoOutput
func SetNoOutput(b bool) {
	options.NoOutput = b
}

// GetNoOutput
func GetNoOutput() bool {
	return options.NoOutput
}

// GetSuppressWarning
func GetSuppressWarning() bool {
	return options.SuppressWarnings
}

// SetConflictsAsErrors
func SetConflictsAsErrors(b bool) {
	options.ConflictsAreErrors = b
}

// GetConflictsAsErrors
func GetConflictsAsErrors() bool {
	return options.ConflictsAreErrors
}

// SetRate
func SetRate(r Rate) {
	apiRate = r
}

// GetRate
func GetRate() Rate {
	return apiRate
}

// SetAPI
func SetAPI(a API) {
	// prod is the default
	if a == "" {
		options.Api = PROD
	} else {
		options.Api = a
	}
}

// GetAPI
func GetAPI() API {
	return options.Api
}

// GetMetadataToken
func GetMetadataToken() bool {
	return options.MetadataToken
}

// SetMetadataToken
func SetMetadataToken(b bool) {
	options.MetadataToken = b
}
