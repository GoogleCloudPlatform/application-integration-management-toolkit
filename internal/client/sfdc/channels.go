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
	"fmt"
	"internal/apiclient"
	"net/url"
	"path"
	"strings"
)

type channel struct {
	Name         string `json:"name,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Description  string `json:"description,omitempty"`
	ChannelTopic string `json:"channelTopic,omitempty"`
	IsActive     bool   `json:"isActive,omitempty"`
	UpateTime    string `json:"upateTime,omitempty"`
	CreateTime   string `json:"createTime,omitempty"`
	DeleteTime   string `json:"deleteTime,omitempty"`
	LastReplayId string `json:"lastReplayId,omitempty"`
}

type channels struct {
	SfdcChannels []channel `json:"sfdcChannels,omitempty"`
}

type channelExternal struct {
	DisplayName  string `json:"displayName,omitempty"`
	Description  string `json:"description,omitempty"`
	ChannelTopic string `json:"channelTopic,omitempty"`
}

// CreateChannelFromContent
func CreateChannelFromContent(instanceVersion string, content []byte) apiclient.APIResponse {
	c := channel{}

	if err := json.Unmarshal(content, &c); err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances")

	u.Path = path.Join(u.Path, "sfdcInstances", instanceVersion, "sfdcChannels")
	return apiclient.HttpClient(u.String(), string(content))
}

// CreateChannel
func CreateChannel(name string, instance string, description string, channelTopic string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())

	channelStr := []string{}

	channelStr = append(channelStr, "\"displayName\":\""+name+"\"")
	channelStr = append(channelStr, "\"description\":\""+description+"\"")
	channelStr = append(channelStr, "\"channelTopic\":\""+channelTopic+"\"")

	payload := "{" + strings.Join(channelStr, ",") + "}"

	u.Path = path.Join(u.Path, "sfdcInstances", instance, "sfdcChannels")
	return apiclient.HttpClient(u.String(), payload)
}

// GetChannel
func GetChannel(name string, instance string, minimal bool) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances", instance, "sfdcChannels", name)

	response := apiclient.HttpClient(u.String())
	if response.Err != nil {
		return response
	}

	if minimal {
		iversion := channel{}
		err := json.Unmarshal(response.RespBody, &iversion)
		if err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      err,
			}
		}
		eversion := convertInternalChannelToExternal(iversion)
		response.RespBody, response.Err = json.Marshal(eversion)
		return response
	}
	return response
}

// ListChannels
func ListChannels(instance string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances", instance, "sfdcChannels")
	return apiclient.HttpClient(u.String())
}

// FindChannel
func FindChannel(name string, instance string) (version string, response apiclient.APIResponse) {
	clist := channels{}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances", instance, "sfdcChannels")
	if response = apiclient.HttpClient(u.String()); response.Err != nil {
		return "", apiclient.APIResponse{
			RespBody: nil,
			Err:      response.Err,
		}
	}

	if err := json.Unmarshal(response.RespBody, &clist); err != nil {
		return "", apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	for _, c := range clist.SfdcChannels {
		if c.DisplayName == name {
			version = c.Name[strings.LastIndex(c.Name, "/")+1:]
			respBody, err := json.Marshal(&c)
			return version, apiclient.APIResponse{
				RespBody: respBody,
				Err:      err,
			}
		}
	}
	return "", apiclient.APIResponse{
		RespBody: nil,
		Err:      fmt.Errorf("channel not found"),
	}
}

// GetInstancesAndChannels
func GetInstancesAndChannels(instances map[string]string) (instancesContent map[string]string, err error) {
	instancesContent = make(map[string]string)

	for instance, channel := range instances {
		instanceUuid, instancesResp, err := FindInstance(instance)
		if err != nil {
			return instancesContent, err
		}
		_, response := FindChannel(channel, instanceUuid)
		if response.Err != nil {
			return instancesContent, err
		}
		instancesContent[string(instancesResp)] = string(response.RespBody)
	}
	return instancesContent, err
}

// convertInternalChannelToExternal
func convertInternalChannelToExternal(internalVersion channel) (externalVersion channelExternal) {
	externalVersion = channelExternal{}

	externalVersion.DisplayName = internalVersion.Name
	externalVersion.Description = internalVersion.Description
	externalVersion.ChannelTopic = internalVersion.ChannelTopic

	return externalVersion
}
