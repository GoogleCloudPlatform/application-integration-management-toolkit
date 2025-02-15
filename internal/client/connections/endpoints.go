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
	"fmt"
	"internal/apiclient"
	"internal/clilog"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type endpoints struct {
	EndpointAttachments []endpoint `json:"endpointAttachments,omitempty"`
	NextPageToken       string     `json:"nextPageToken,omitempty"`
}

type endpoint struct {
	Name              string `json:"name,omitempty"`
	CreateTime        string `json:"createTime,omitempty"`
	UpdateTime        string `json:"updateTime,omitempty"`
	ServiceAttachment string `json:"serviceAttachment,omitempty"`
	EndpointIP        string `json:"endpointIp,omitempty"`
}

type endpointExternal struct {
	ServiceAttachment string `json:"serviceAttachment,omitempty"`
}

// CreateEndpoint
func CreateEndpoint(name string, serviceAttachment string, description string, wait bool) apiclient.APIResponse {
	endpointStr := []string{}

	endpointStr = append(endpointStr, "\"name\":\""+
		fmt.Sprintf("projects/%s/locations/%s/endpointAttachments/%s",
			apiclient.GetProjectID(), apiclient.GetRegion(), name)+"\"")
	endpointStr = append(endpointStr, "\"serviceAttachment\":\""+serviceAttachment+"\"")

	if description != "" {
		endpointStr = append(endpointStr, "\"description\":\""+description+"\"")
	}

	payload := "{" + strings.Join(endpointStr, ",") + "}"

	u, _ := url.Parse(apiclient.GetBaseConnectorEndpointAttachURL())
	u.Path = path.Join(u.Path)

	q := u.Query()
	q.Set("endpointAttachmentId", name)
	u.RawQuery = q.Encode()

	response := apiclient.HttpClient(u.String(), payload)
	if response.Err != nil {
		return response
	}

	if wait {

		o := operation{}
		if err := json.Unmarshal(response.RespBody, &o); err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      err,
			}
		}

		operationId := filepath.Base(o.Name)
		clilog.Info.Printf("Checking connection status for %s in %d seconds\n", operationId, interval)

		stop := apiclient.Every(interval*time.Second, func(time.Time) bool {
			var respBody []byte

			if response := GetOperation(operationId); response.Err != nil {
				return false
			}

			if err := json.Unmarshal(respBody, &o); err != nil {
				return false
			}

			if o.Done {
				if o.Error != nil {
					clilog.Error.Printf("Connection completed with error: %s\n", o.Error.Message)
				} else {
					clilog.Info.Println("Connection completed successfully!")
				}
				return false
			} else {
				clilog.Info.Printf("Connection status is: %t. Waiting %d seconds.\n", o.Done, interval)
				return true
			}
		})

		<-stop
	}
	return response
}

// GetEndpoint
func GetEndpoint(name string, overrides bool) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorEndpointAttachURL())
	u.Path = path.Join(u.Path, name)

	response := apiclient.HttpClient(u.String())
	if response.Err != nil {
		return response
	}

	if overrides {
		e := endpoint{}
		if err := json.Unmarshal(response.RespBody, &e); err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      err,
			}
		}
		eversion := convertInternalToExternal(e)
		respBody, err := json.Marshal(eversion)
		if err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      err,
			}
		}
		return apiclient.APIResponse{
			RespBody: respBody,
			Err:      nil,
		}
	}
	return response
}

// ListEndpoints
func ListEndpoints(pageSize int, pageToken string, filter string, orderBy string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorEndpointAttachURL())
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

// DeleteEndpoint
func DeleteEndpoint(name string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorEndpointAttachURL())
	u.Path = path.Join(u.Path, name)
	return apiclient.HttpClient(u.String(), "", "DELETE")
}

func FindEndpoint(name string) (found bool) {
	var pageToken string
	var respBody []byte

	for {
		if response := ListEndpoints(maxPageSize, pageToken, "", ""); response.Err != nil {
			return false
		}
		l := endpoints{}
		if err := json.Unmarshal(respBody, &l); err != nil {
			return false
		}
		for _, e := range l.EndpointAttachments {
			if e.Name[strings.LastIndex(e.Name, "/")+1:] == name {
				return true
			}
		}
		if l.NextPageToken != "" {
			pageToken = l.NextPageToken
			continue
		} else {
			return false
		}
	}
}

// convertInternalToExternal
func convertInternalToExternal(internalVersion endpoint) (externalVersion endpointExternal) {
	externalVersion = endpointExternal{}
	externalVersion.ServiceAttachment = internalVersion.ServiceAttachment
	return externalVersion
}
