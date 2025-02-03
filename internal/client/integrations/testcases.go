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

package integrations

import (
	"encoding/json"
	"fmt"
	"internal/apiclient"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
)

type testCase struct {
	Name                      string              `json:"name,omitempty"`
	Description               string              `json:"description,omitempty"`
	DisplayName               string              `json:"displayName,omitempty"`
	TriggerId                 string              `json:"triggerId,omitempty"`
	TestTaskConfigs           []testTaskConfig    `json:"testTaskConfigs,omitempty"`
	DatabasePersistencePolicy *string             `json:"databasePersistencePolicy,omitempty"`
	TriggerConfig             triggerconfig       `json:"triggerConfig,omitempty"`
	TestInputParameters       []parameterExternal `json:"testInputParameters,omitempty"`
}

type listTestCases struct {
	TestCases     []testCase `json:"testCases,omitempty"`
	NextPageToken string     `json:"nextPageToken,omitempty"`
}

type testTaskConfig struct {
	TaskNumber string      `json:"taskNumber,omitempty"`
	Task       string      `json:"task,omitempty"`
	TaskConfig taskconfig  `json:"taskConfig,omitempty"`
	Assertions []assertion `json:"assertions,omitempty"`
	MockConfig mockConfig  `json:"mockConfig,omitempty"`
}

type assertion struct {
	AssertionStrategy string          `json:"assertionStrategy,omitempty"`
	Parameter         *eventparameter `json:"parameter,omitempty"`
	Condition         string          `json:"condition,omitempty"`
	RetryCount        int             `json:"retryCount,omitempty"`
}

type mockConfig struct {
	MockStrategy     string           `json:"mockStrategy,omitempty"`
	Parameters       []eventparameter `json:"parameters,omitempty"`
	FailedExecutions string           `json:"failedExecutions,omitempty"`
}

type testCaseResponse struct {
	ExecutionId        string            `json:"executionId,omitempty`
	OutputParameters   interface{}       `json:"outputParameters,omitempty`
	AssertionResults   []assertionResult `json:"assertionResults,omitempty`
	TestExecutionState string            `json:"testExecutionState,omitempty`
}

type assertionResult struct {
	TaskNumber     string      `json:"taskNumber,omitempty"`
	Assertion      interface{} `json:"assertion,omitempty"`
	TaskName       string      `json:"taskName,omitempty"`
	Status         string      `json:"status,omitempty"`
	FailureMessage string      `json:"failureMessage,omitempty"`
}

func CreateTestCase(name string, version string, content string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases")
	return apiclient.HttpClient(u.String(), content)
}

func CreateTestCaseBySnapshot(name string, snapshot string, content string) apiclient.APIResponse {
	version, err := getTestCaseIntegrationVersion(name, snapshot, "")
	if err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}
	return CreateTestCase(name, version, content)
}

func CreateTestCaseByUserLabel(name string, userLabel string, content string) apiclient.APIResponse {
	version, err := getTestCaseIntegrationVersion(name, "", userLabel)
	if err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}
	return CreateTestCase(name, version, content)
}

func DeleteTestCase(name string, version string, testCaseID string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases", testCaseID)
	return apiclient.HttpClient(u.String(), "", "DELETE")
}

func GetTestCase(name string, version string, testCaseID string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases", testCaseID)
	return apiclient.HttpClient(u.String())
}

func ListTestCases(name string, version string, full bool, filter string,
	pageSize int, pageToken string, orderBy string) apiclient.APIResponse {

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
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
	if orderBy != "" {
		q.Set("orderBy", orderBy)
	}

	u.RawQuery = q.Encode()
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases")
	response := apiclient.HttpClient(u.String())
	if !full {
		return getTestCases(response.RespBody, full)
	}
	return response
}

func ExecuteTestCase(name string, version string, testCaseID string, content string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases", testCaseID, ":executeTest")
	return apiclient.HttpClient(u.String(), content)
}

func AssertTestExecutionResult(testBody []byte) error {
	tr := testCaseResponse{}
	err := json.Unmarshal(testBody, &tr)
	if err != nil {
		return err
	}
	if tr.TestExecutionState == "PASSED" {
		return nil
	}
	return fmt.Errorf("test failed with %d assertions", len(tr.AssertionResults))
}

func ListTestCasesByUserlabel(name string, userLabel string, full bool, filter string,
	pageSize int, pageToken string, orderBy string) apiclient.APIResponse {

	version, err := getTestCaseIntegrationVersion(name, "", userLabel)
	if err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}
	return ListTestCases(name, version, full, filter, pageSize, pageToken, orderBy)
}

func ListTestCasesBySnapshot(name string, snapshot string, full bool, filter string,
	pageSize int, pageToken string, orderBy string) apiclient.APIResponse {

	version, err := getTestCaseIntegrationVersion(name, snapshot, "")
	if err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}
	return ListTestCases(name, version, full, filter, pageSize, pageToken, orderBy)
}

// FindTestCase
func FindTestCase(name string, integrationVersion string, displayName string, pageToken string) (version string, err error) {
	lt := listTestCases{}
	var respBody []byte

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	if pageToken != "" {
		q := u.Query()
		q.Set("pageToken", pageToken)
		u.RawQuery = q.Encode()
	}

	u.Path = path.Join(u.Path, "integrations", name, "versions", integrationVersion, "testCases")
	if response := apiclient.HttpClient(u.String()); response.Err != nil {
		return "", response.Err
	}

	if err = json.Unmarshal(respBody, &lt); err != nil {
		return "", err
	}

	for _, testcase := range lt.TestCases {
		if testcase.DisplayName == displayName {
			version = filepath.Base(testcase.Name)
			return version, nil
		}
	}
	if lt.NextPageToken != "" {
		return FindTestCase(name, integrationVersion, displayName, lt.NextPageToken)
	}
	return "", fmt.Errorf("testCase not found")
}

func ListAllTestCases(name string, version string) apiclient.APIResponse {
	l := listTestCases{}
	var err error
	var response apiclient.APIResponse

	for {
		newltc := listTestCases{}
		response := ListTestCases(name, version, true, "", -1, "", "")
		if response.Err != nil {
			return response
		}
		err = json.Unmarshal(response.RespBody, &newltc)
		if err != nil {
			return apiclient.APIResponse{
				Err: err,
			}
		}

		l.TestCases = append(l.TestCases, newltc.TestCases...)

		if newltc.NextPageToken == "" {
			break
		}
	}

	response.RespBody, response.Err = json.Marshal(l)

	return response
}

func DeleteAllTestCases(name string, version string) (err error) {
	response := ListAllTestCases(name, version)
	if response.Err != nil {
		return response.Err
	}

	l := listTestCases{}
	err = json.Unmarshal(response.RespBody, &l)
	if err != nil {
		return err
	}

	for _, tc := range l.TestCases {
		response := DeleteTestCase(name, version, filepath.Base(tc.Name))
		if response.Err != nil {
			return response.Err
		}
	}

	return nil
}

func getTestCaseIntegrationVersion(name string, snapshot string, userLabel string) (version string, err error) {

	var response apiclient.APIResponse

	if snapshot != "" {
		response = GetBySnapshot(name, snapshot, false, false, false)
		if response.Err != nil {
			return "", response.Err
		}
	} else {
		response = GetByUserlabel(name, userLabel, false, false, false)
		if response.Err != nil {
			return "", response.Err
		}
	}

	iversion := integrationVersion{}
	err = json.Unmarshal(response.RespBody, &iversion)
	if err != nil {
		return "", err
	}

	version = getVersion(iversion.Name)

	return version, nil

}

func getTestCases(respBody []byte, full bool) apiclient.APIResponse {

	var response apiclient.APIResponse

	if full {
		return apiclient.APIResponse{
			RespBody: respBody,
			Err:      nil,
		}
	}

	l := listTestCases{}
	newltc := listTestCases{}

	err := json.Unmarshal(respBody, &l)
	if err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}
	for _, tc := range l.TestCases {
		if tc.DatabasePersistencePolicy != nil && *tc.DatabasePersistencePolicy == "" {
			tc.DatabasePersistencePolicy = nil
		}
		newltc.TestCases = append(newltc.TestCases, tc)
	}
	response.RespBody, response.Err = json.Marshal(newltc.TestCases)
	return response
}
