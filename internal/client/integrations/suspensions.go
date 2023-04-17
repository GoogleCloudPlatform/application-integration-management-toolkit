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
	"net/url"
	"path"
	"strconv"

	"internal/apiclient"
)

// List all suspensions
func ListSuspensions(name string, execution string, pageSize int, pageToken string, filter string, orderBy string) (respBody []byte, err error) {
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
	u.Path = path.Join(u.Path, "integrations", name, "executions", execution, "suspensions")
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

// Lift a suspension
func Lift(name string, execution string, suspension string, result string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "executions", execution, "suspensions", suspension, ":lift")
	payload := "{ \"suspension_result\":\"" + result + "\"}"
	respBody, err = apiclient.HttpClient(u.String(), payload)
	return respBody, err
}

// Resolve one or more suspensions
func Resolve(name string, suspensions string) (respBody []byte, err error) {
	return nil, nil
}
