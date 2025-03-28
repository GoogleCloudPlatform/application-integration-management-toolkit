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
	"encoding/json"
	"internal/apiclient"
	"net/url"
	"path"
	"strconv"
)

type zone struct {
	DNS           string `json:"dns,omitempty"`
	Description   string `json:"desciption,omitempty"`
	TargetProject string `json:"targetProject,omitempty"`
	TargetVPC     string `json:"targetVpc,omitempty"`
}

// CreateZone
func CreateZone(name string, content []byte) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorZonesURL())
	q := u.Query()
	q.Set("managedZoneId", name)
	u.RawQuery = q.Encode()

	z := zone{}
	if err := json.Unmarshal(content, &z); err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	u.Path = path.Join(u.Path, name)
	return apiclient.HttpClient(u.String(), string(content))
}

// GetZone
func GetZone(name string, overrides bool) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorZonesURL())
	u.Path = path.Join(u.Path, name)

	response := apiclient.HttpClient(u.String())

	if overrides {
		z := zone{}
		if err := json.Unmarshal(response.RespBody, &z); err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      err,
			}
		}
		response.RespBody, response.Err = json.Marshal(z)
		return response
	}
	return response
}

// DeleteZone
func DeleteZone(name string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorZonesURL())
	u.Path = path.Join(u.Path, name)
	return apiclient.HttpClient(u.String(), "", "DELETE")
}

// ListZones
func ListZones(pageSize int, pageToken string, filter string, orderBy string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorZonesURL())
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
