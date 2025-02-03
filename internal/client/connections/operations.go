// Copyright 2023 Google LLC
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
	"internal/apiclient"
	"net/url"
	"path"
	"strconv"
)

// GetOperation
func GetOperation(name string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorOperationsrURL())
	u.Path = path.Join(u.Path, name)
	return apiclient.HttpClient(u.String())
}

// ListOperations
func ListOperations(pageSize int, pageToken string, filter string, orderBy string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorOperationsrURL())
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
	return apiclient.HttpClient(u.String())
}

// CancelOperation
func CancelOperation(name string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorOperationsrURL())
	u.Path = path.Join(u.Path, name+":cancel")
	return apiclient.HttpClient(u.String(), "")
}
