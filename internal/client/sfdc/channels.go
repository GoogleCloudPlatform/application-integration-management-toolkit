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
	"net/url"
	"path"
	"strings"

	"internal/apiclient"
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
func CreateChannelFromContent(instanceVersion string, content []byte) (respBody []byte, err error) {
	c := channel{}

	if err = json.Unmarshal(content, &c); err != nil {
		return nil, err
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances")

	u.Path = path.Join(u.Path, "sfdcInstances", instanceVersion, "sfdcChannels")
	respBody, err = apiclient.HttpClient(u.String(), string(content))
	return respBody, err
}

// CreateChannel
func CreateChannel(name string, instance string, description string, channelTopic string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())

	channelStr := []string{}

	channelStr = append(channelStr, "\"displayName\":\""+name+"\"")
	channelStr = append(channelStr, "\"description\":\""+description+"\"")
	channelStr = append(channelStr, "\"channelTopic\":\""+channelTopic+"\"")

	payload := "{" + strings.Join(channelStr, ",") + "}"

	u.Path = path.Join(u.Path, "sfdcInstances", instance, "sfdcChannels")
	respBody, err = apiclient.HttpClient(u.String(), payload)
	return respBody, err
}

// GetChannel
func GetChannel(name string, instance string, minimal bool) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances", instance, "sfdcChannels", name)

	if minimal {
		apiclient.ClientPrintHttpResponse.Set(false)
	}
	respBody, err = apiclient.HttpClient(u.String())
	if minimal {
		iversion := channel{}
		err := json.Unmarshal(respBody, &iversion)
		if err != nil {
			return nil, err
		}
		eversion := convertInternalChannelToExternal(iversion)
		respBody, err = json.Marshal(eversion)
		if err != nil {
			return nil, err
		}
		apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
		apiclient.PrettyPrint(respBody)

	}
	return respBody, err
}

// ListChannels
func ListChannels(instance string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances", instance, "sfdcChannels")
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

// FindChannel
func FindChannel(name string, instance string) (version string, respBody []byte, err error) {
	clist := channels{}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "sfdcInstances", instance, "sfdcChannels")
	if respBody, err = apiclient.HttpClient(u.String()); err != nil {
		return "", nil, err
	}

	if err = json.Unmarshal(respBody, &clist); err != nil {
		return "", nil, err
	}

	for _, c := range clist.SfdcChannels {
		if c.DisplayName == name {
			version = c.Name[strings.LastIndex(c.Name, "/")+1:]
			respBody, err := json.Marshal(&c)
			return version, respBody, err
		}
	}
	return "", nil, fmt.Errorf("instance not found")
}

// GetInstancesAndChannels
func GetInstancesAndChannels(instances map[string]string) (instancesContent map[string]string, err error) {
	instancesContent = make(map[string]string)

	for instance, channel := range instances {
		instanceUuid, instancesResp, err := FindInstance(instance)
		if err != nil {
			return instancesContent, err
		}
		_, channelResp, err := FindChannel(channel, instanceUuid)
		if err != nil {
			return instancesContent, err
		}
		instancesContent[string(instancesResp)] = string(channelResp)
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
