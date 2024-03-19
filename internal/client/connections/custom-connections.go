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

package connections

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"internal/apiclient"
)

type customConnectorOverrides struct {
	DisplayName            string                        `json:"displayName,omitempty"`
	Description            string                        `json:"description,omitempty"`
	CustomConnectorType    string                        `json:"customConnectorType,omitempty"`
	Labels                 map[string]string             `json:"labels,omitempty"`
	CustomConnectorVersion customConnectorVersionRequest `json:"customConnectorVersion,omitempty"`
}

type customConnectorVersionRequest struct {
	Labels                         map[string]string        `json:"labels,omitempty"`
	ServiceAccount                 *string                  `json:"serviceAccount,omitempty"`
	EnableBackendDestinationConfig bool                     `json:"enableBackendDestinationConfig,omitempty"`
	SpecLocation                   string                   `json:"specLocation,omitempty"`
	AuthConfig                     *authConfig              `json:"authConfig,omitempty"`
	DestinationConfigs             []destinationConfig      `json:"destinationConfigs,omitempty"`
	BackendVariableTemplates       []configVariableTemplate `json:"backendVariableTemplates,omitempty"`
}

type configVariableTemplate struct {
	Key             string `json:"key,omitempty"`
	ValueType       string `json:"valueType,omitempty"`
	DisplayName     string `json:"displayName,omitempty"`
	Description     string `json:"description,omitempty"`
	ValidationRegex string `json:"validationRegex,omitempty"`
	Required        bool   `json:"required,omitempty"`
	IsAdvanced      bool   `json:"isAdvanced,omitempty"`
	LocationType    string `json:"locationType,omitempty"`
}

// CreateCustom
func CreateCustom(name string, description string, displayName string,
	connType string, labels map[string]string,
) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseCustomConnectorURL())
	q := u.Query()
	q.Set("customConnectorId", name)

	customConnect := []string{}

	customConnect = append(customConnect, "\"displayName\":"+"\""+displayName+"\"")
	customConnect = append(customConnect, "\"description\":"+"\""+description+"\"")
	customConnect = append(customConnect, "\"customConnectorType\":"+"\""+connType+"\"")

	if len(labels) > 0 {
		l := []string{}
		for key, value := range labels {
			l = append(l, "\""+key+"\":\""+value+"\"")
		}
		labelStr := "\"labels\":{" + strings.Join(l, ",") + "}"
		customConnect = append(customConnect, labelStr)
	}

	payload := "{" + strings.Join(customConnect, ",") + "}"
	u.RawQuery = q.Encode()
	respBody, err = apiclient.HttpClient(u.String(), payload)
	return respBody, err
}

// DeleteCustom
func DeleteCustom(name string, force bool) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseCustomConnectorURL())
	u.Path = path.Join(u.Path, name)
	q := u.Query()
	q.Set("force", strconv.FormatBool(force))
	u.RawQuery = q.Encode()
	respBody, err = apiclient.HttpClient(u.String(), "", "DELETE")
	return respBody, err
}

// GetCustom
func GetCustom(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseCustomConnectorURL())
	u.Path = path.Join(u.Path, name)
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

// ListCustom
func ListCustom(pageSize int, pageToken string, filter string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseCustomConnectorURL())
	q := u.Query()
	if pageSize != -1 {
		q.Set("pageSize", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		q.Set("pageToken", pageToken)
	}
	if filter != "" {
		q.Set("filter", filter)
	}
	u.RawQuery = q.Encode()
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

// CreateCustomVersion
func CreateCustomVersion(connName string, versionName string, content []byte,
	serviceAccountName string, serviceAccountProject string,
) (respBody []byte, err error) {
	c := customConnectorVersionRequest{}
	if err = json.Unmarshal(content, &c); err != nil {
		return nil, err
	}

	// service account overrides have been provided, use them
	if serviceAccountName != "" {
		// set the project id if one was not presented
		if serviceAccountProject == "" {
			serviceAccountProject = apiclient.GetProjectID()
		}
		serviceAccountName = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", serviceAccountName, serviceAccountProject)
		// create the SA if it doesn't exist
		if err = apiclient.CreateServiceAccount(serviceAccountName); err != nil {
			return nil, err
		}
	}

	if c.ServiceAccount != nil && serviceAccountName != "" {
		*c.ServiceAccount = serviceAccountName
	}

	if content, err = json.Marshal(c); err != nil {
		return nil, err
	}

	u, _ := url.Parse(apiclient.GetBaseCustomConnectorURL())
	u.Path = path.Join(u.Path, connName, "customConnectorVersions")
	q := u.Query()
	q.Set("customConnectorVersionId", versionName)
	u.RawQuery = q.Encode()

	respBody, err = apiclient.HttpClient(u.String(), string(content))
	return respBody, err
}

func GetCustomVersion(connName string, connVersion string, overrides bool) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseCustomConnectorURL())
	u.Path = path.Join(u.Path, connName, "customConnectorVersions", connVersion)
	if overrides {
		apiclient.ClientPrintHttpResponse.Set(false)
		c := customConnectorOverrides{}
		connRespBody, err := GetCustom(connName)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(connRespBody, &c); err != nil {
			return nil, err
		}
		respBody, err = apiclient.HttpClient(u.String())
		if err != nil {
			return nil, err
		}
		apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
		cVerReq := customConnectorVersionRequest{}
		if err = json.Unmarshal(respBody, &cVerReq); err != nil {
			return nil, err
		}
		// remove the default p4s from the overrides
		if strings.Contains(*cVerReq.ServiceAccount, "-compute@developer.gserviceaccount.com") {
			cVerReq.ServiceAccount = nil
		}
		c.CustomConnectorVersion = cVerReq
		overridesResp, err := json.Marshal(c)
		if err != nil {
			return nil, err
		}
		apiclient.PrettyPrint(overridesResp)
		return overridesResp, nil
	}
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

func ListCustomVersions(connName string, pageSize int, pageToken string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseCustomConnectorURL())
	u.Path = path.Join(u.Path, connName, "customConnectorVersions")
	q := u.Query()
	if pageSize != -1 {
		q.Set("pageSize", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		q.Set("pageToken", pageToken)
	}
	u.RawQuery = q.Encode()
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

func GetCustomFromConnection(contents []byte) (respBody []byte, err error) {
	c := connection{}
	err = json.Unmarshal(respBody, &c)
	if err != nil {
		return nil, err
	}
	if c.ConnectorDetails.Provider != "customconnector" {
		return nil, fmt.Errorf("connector is not of type customconnector")
	}
	respBody, err = GetCustomVersion(getConnectorName(*c.ConnectorVersion), getConnectorVersionId(*c.ConnectorVersion), false)
	return respBody, err
}

func IsCustomConnector(contents []byte) bool {
	c := connection{}
	err := json.Unmarshal(contents, &c)
	if err != nil {
		return false
	}
	if c.ConnectorDetails.Provider != "customconnector" {
		return false
	}
	return true
}

func CreateCustomWithVersion(name string, version string, contents []byte,
	serviceAccount string, serviceAccountProject string) (err error) {
	c := customConnectorOverrides{}
	err = json.Unmarshal(contents, &c)
	if err != nil {
		return err
	}
	createCustomBody, err := CreateCustom(name, c.Description, c.DisplayName, c.CustomConnectorType, c.Labels)
	if err != nil {
		return err
	}

	var createCustomMap map[string]interface{}
	err = json.Unmarshal(createCustomBody, &createCustomMap)
	if err != nil {
		return err
	}

	// wait for custom connection to be created
	operationName := strings.Split(fmt.Sprintf("%s", createCustomMap["name"]), "/")[5]
	err = waitForCustom(operationName)
	if err != nil {
		return err
	}

	connectionVersionContents, err := json.Marshal(c.CustomConnectorVersion)
	if err != nil {
		return err
	}
	_, err = CreateCustomVersion(name, version, connectionVersionContents, serviceAccount, serviceAccountProject)
	if err != nil {
		return err
	}

	// wait for custom version to be created
	err = waitForCustomVersion(name, version)
	return err
}

func waitForCustom(operationName string) error {
	var err error
	var respBody []byte
	var respMap map[string]interface{}

	region := apiclient.GetRegion()
	defer apiclient.SetRegion(region)

	apiclient.SetRegion("global")

	for {
		if respBody, err = GetOperation(operationName); err != nil {
			return err
		}
		if err = json.Unmarshal(respBody, &respMap); err != nil {
			return err
		}
		done := respMap["done"].(bool)
		if done {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}

func waitForCustomVersion(name string, version string) error {
	var err error
	var respBody []byte
	var respMap map[string]string

	for {
		if respBody, err = GetCustomVersion(name, version, false); err != nil {
			return err
		}

		if err = json.Unmarshal(respBody, &respMap); err != nil {
			return err
		}

		if respMap["state"] == "ACTIVE" {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}
