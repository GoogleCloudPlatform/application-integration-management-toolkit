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
	"internal/apiclient"
	"net/url"
	"path"
)

func CreateTestCase(name string, version string, content string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases")
	respBody, err = apiclient.HttpClient(u.String(), content)
	return respBody, err
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

func ListTestCases(name string, version string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases")
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

func RunTestCase(name string, version string, testCaseID string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version, "testCases", testCaseID, ":executeTest")
	respBody, err = apiclient.HttpClient(u.String(), "")
	return respBody, err
}
