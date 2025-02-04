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

package sfdc

import (
	"encoding/json"
	"internal/apiclient"
	"net/url"
	"path"
	"strings"
)

type instance struct {
	Name             string   `json:"name,omitempty"`
	DisplayName      string   `json:"displayName,omitempty"`
	Description      string   `json:"description,omitempty"`
	SfdcOrgId        string   `json:"sfdcOrgId,omitempty"`
	AuthConfigId     []string `json:"authConfigId,omitempty"`
	UpateTime        string   `json:"upateTime,omitempty"`
	CreateTime       string   `json:"createTime,omitempty"`
	DeleteTime       string   `json:"deleteTime,omitempty"`
	ServiceAuthority string   `json:"serviceAuthority,omitempty"`
}

type instances struct {
	SfdcInstances []instance `json:"sfdcInstances,omitempty"`
	NextPageToken string     `json:"nextPageToken,omitempty"`
}

type instanceExternal struct {
	DisplayName      string   `json:"displayName,omitempty"`
	Description      string   `json:"description,omitempty"`
	SfdcOrgId        string   `json:"sfdcOrgId,omitempty"`
	AuthConfigId     []string `json:"authConfigId,omitempty"`
	ServiceAuthority string   `json:"serviceAuthority,omitempty"`
}

// CreateInstanceFromContent
func CreateInstanceFromContent(content []byte) apiclient.APIResponse {
	i := instance{}

	if err := json.Unmarshal(content, &i); err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      apiclient.NewCliError("error unmarshalling", err),
		}
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances")

	return apiclient.HttpClient(u.String(), string(content))
}

// CreateInstance
func CreateInstance(name string, description string, sfdcOrgId string, serviceAuthority string, authConfig []string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())

	if len(authConfig) < 1 {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      apiclient.NewCliError("at least one authConfig must be sent", nil),
		}
	}

	instanceStr := []string{}
	instanceStr = append(instanceStr, "\"displayName\":\""+name+"\"")
	instanceStr = append(instanceStr, "\"description\":\""+description+"\"")
	instanceStr = append(instanceStr, "\"sfdcOrgId\":\""+sfdcOrgId+"\"")
	instanceStr = append(instanceStr, "\"serviceAuthority\":\""+serviceAuthority+"\"")

	authConfigsStr := "\"attributes\":[" + strings.Join(authConfig, ",") + "]"

	instanceStr = append(instanceStr, "\"authConfigId\":\""+authConfigsStr+"\"")

	payload := "{" + strings.Join(instanceStr, ",") + "}"
	u.Path = path.Join(u.Path, "sfdcInstances")
	return apiclient.HttpClient(u.String(), payload)
}

// GetInstance
func GetInstance(name string, minimal bool) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances", name)

	response := apiclient.HttpClient(u.String())
	if response.Err != nil {
		return response
	}

	if minimal {
		iversion := instance{}
		err := json.Unmarshal(response.RespBody, &iversion)
		if err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      apiclient.NewCliError("error unmarshalling", err),
			}
		}
		eversion := convertInternalInstanceToExternal(iversion)
		response.RespBody, response.Err = json.Marshal(eversion)
		return response
	}
	return response
}

// ListInstances
func ListInstances() apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances")
	return apiclient.HttpClient(u.String())
}

// FindInstance
func FindInstance(name string) (version string, respBody []byte, err error) {
	ilist := instances{}

	response := ListInstances()
	if response.Err != nil {
		return "", response.RespBody, response.Err
	}
	if err = json.Unmarshal(response.RespBody, &ilist); err != nil {
		return "", nil, apiclient.NewCliError("error unmarshalling", err)
	}

	for _, i := range ilist.SfdcInstances {
		if i.DisplayName == name {
			version = i.Name[strings.LastIndex(i.Name, "/")+1:]
			respBody, err := json.Marshal(&i)
			return version, respBody, apiclient.NewCliError("error marshallilng", err)
		}
	}
	return "", nil, apiclient.NewCliError("instance not found", nil)
}

// convertInternalInstanceToExternal
func convertInternalInstanceToExternal(internalVersion instance) (externalVersion instanceExternal) {
	externalVersion = instanceExternal{}

	externalVersion.DisplayName = internalVersion.Name
	externalVersion.Description = internalVersion.Description
	externalVersion.ServiceAuthority = internalVersion.ServiceAuthority
	externalVersion.SfdcOrgId = internalVersion.SfdcOrgId
	externalVersion.AuthConfigId = internalVersion.AuthConfigId

	return externalVersion
}
