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
	"internal/apiclient"
	"net/url"
	"path"
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
	AssertionStrategy string         `json:"assertionStrategy,omitempty"`
	Parameter         eventparameter `json:"parameter,omitempty"`
	Condition         string         `json:"condition,omitempty"`
	RetryCount        int            `json:"retryCount,omitempty"`
}

type mockConfig struct {
	MockStrategy     string           `json:"mockStrategy,omitempty"`
	Parameters       []eventparameter `json:"parameters,omitempty"`
	FailedExecutions string           `json:"failedExecutions,omitempty"`
}

func CreateTestCase(name string, version string, content string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases")
	respBody, err = apiclient.HttpClient(u.String(), content)
	return respBody, err
}

func CreateTestCaseBySnapshot(name string, snapshot string, content string) (respBody []byte, err error) {
	version, err := getTestCaseIntegrationVersion(name, snapshot, "")
	if err != nil {
		return nil, err
	}
	return CreateTestCase(name, version, content)
}

func CreateTestCaseByUserLabel(name string, userLabel string, content string) (respBody []byte, err error) {
	version, err := getTestCaseIntegrationVersion(name, "", userLabel)
	if err != nil {
		return nil, err
	}
	return CreateTestCase(name, version, content)
}

func DeleteTestCase(name string, version string, testCaseID string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases", testCaseID)
	respBody, err = apiclient.HttpClient(u.String(), "", "DELETE")
	return respBody, err
}

func GetTestCase(name string, version string, testCaseID string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases", testCaseID)
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

func ListTestCases(name string, version string, full bool) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases")
	respBody, err = apiclient.HttpClient(u.String())
	if !full {
		return getTestCases(respBody, full)
	}
	return respBody, err
}

func ExecuteTestCase(name string, version string, testCaseID string, content string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases", testCaseID, ":executeTest")
	respBody, err = apiclient.HttpClient(u.String(), content)
	return respBody, err
}

func ListTestCasesByUserlabel(name string, userLabel string, full bool) (respBody []byte, err error) {

	version, err := getTestCaseIntegrationVersion(name, "", userLabel)
	if err != nil {
		return nil, err
	}
	return ListTestCases(name, version, full)
}

func ListTestCasesBySnapshot(name string, snapshot string, full bool) (respBody []byte, err error) {

	version, err := getTestCaseIntegrationVersion(name, snapshot, "")
	if err != nil {
		return nil, err
	}
	return ListTestCases(name, version, full)
}

func getTestCaseIntegrationVersion(name string, snapshot string, userLabel string) (version string, err error) {

	var iversionBytes []byte

	tmp := apiclient.GetCmdPrintHttpResponseSetting()
	apiclient.DisableCmdPrintHttpResponse()

	if snapshot != "" {
		iversionBytes, err = GetBySnapshot(name, snapshot, false, false, false)
		if err != nil {
			return "", err
		}
	} else {
		iversionBytes, err = GetByUserlabel(name, userLabel, false, false, false)
		if err != nil {
			return "", err
		}
	}

	iversion := integrationVersion{}
	err = json.Unmarshal(iversionBytes, &iversion)
	if err != nil {
		return "", err
	}

	version = getVersion(iversion.Name)

	apiclient.ClientPrintHttpResponse.Set(tmp)
	apiclient.EnableCmdPrintHttpResponse()

	return version, nil

}

func getTestCases(respBody []byte, full bool) (rb []byte, err error) {

	if full {
		return respBody, nil
	}

	l := listTestCases{}
	newltc := listTestCases{}

	err = json.Unmarshal(respBody, &l)
	if err != nil {
		return nil, err
	}
	for _, tc := range l.TestCases {
		if tc.DatabasePersistencePolicy != nil && *tc.DatabasePersistencePolicy == "" {
			tc.DatabasePersistencePolicy = nil
		}
		newltc.TestCases = append(newltc.TestCases, tc)
	}
	return json.Marshal(newltc.TestCases)
}
