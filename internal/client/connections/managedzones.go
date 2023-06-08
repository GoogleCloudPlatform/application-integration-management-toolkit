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
	"net/url"
	"path"
	"strconv"

	"internal/apiclient"
)

type zone struct {
	DNS           string `json:"dns,omitempty"`
	Description   string `json:"desciption,omitempty"`
	TargetProject string `json:"targetProject,omitempty"`
	TargetVPC     string `json:"targetVpc,omitempty"`
}

// CreateZone
func CreateZone(name string, content []byte) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseConnectorZonesURL())
	q := u.Query()
	q.Set("managedZoneId", name)
	u.RawQuery = q.Encode()

	z := zone{}
	err = json.Unmarshal(content, &z)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, name)
	respBody, err = apiclient.HttpClient(u.String(), string(content))
	return respBody, err
}

// GetZone
func GetZone(name string, overrides bool) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseConnectorZonesURL())
	u.Path = path.Join(u.Path, name)
	if overrides {
		apiclient.ClientPrintHttpResponse.Set(false)
	}
	respBody, err = apiclient.HttpClient(u.String())
	if overrides {
		z := zone{}
		if err = json.Unmarshal(respBody, &z); err != nil {
			return nil, err
		}
		return json.Marshal(z)
	}
	return respBody, err
}

// DeleteZone
func DeleteZone(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseConnectorZonesURL())
	u.Path = path.Join(u.Path, name)
	respBody, err = apiclient.HttpClient(u.String(), "", "DELETE")
	return respBody, err
}

// ListZones
func ListZones(pageSize int, pageToken string, filter string, orderBy string) (respBody []byte, err error) {
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
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}
