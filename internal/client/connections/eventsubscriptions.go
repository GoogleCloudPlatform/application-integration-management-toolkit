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

package connections

import (
	"encoding/json"
	"internal/apiclient"
	"net/url"
	"path"
	"strconv"
)

type eventRequest struct {
	Name           string                       `json:"name,omitempty"`
	EventTypeId    string                       `json:"eventTypeId,omitempty"`
	Subscriber     string                       `json:"subscriber,omitempty"`
	SubscriberLink string                       `json:"subscriberLink,omitempty"`
	Destinations   eventSubscriptionDestination `json:"destinations,omitempty"`
	Jms            jms                          `json:"jms,omitempty"`
}

type eventSubscriptionDestination struct {
	Type           string        `json:"type,omitempty"`
	ServiceAccount string        `json:"serviceAccount,omitempty"`
	Endpoint       eventEndpoint `json:"endpoint,omitempty"`
}

type eventEndpoint struct {
	EndpointUri string   `json:"endpointUri,omitempty"`
	Headers     []header `json:"headers,omitempty"`
}

type header struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type jms struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

// CreateEventSubscription
func CreateEventSubscription(connName string, subscriptionId string, contents []byte) apiclient.APIResponse {
	e := eventRequest{}
	if err := json.Unmarshal(contents, &e); err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	u.Path = path.Join(u.Path, "connectors", connName, "eventSubscriptions")
	q := u.Query()
	q.Set("eventSubscriptionId", subscriptionId)
	u.RawQuery = q.Encode()
	return apiclient.HttpClient(u.String(), string(contents))
}

// GetEventSubscription
func GetEventSubscription(name string, connName string, overrides bool) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	u.Path = path.Join(u.Path, "connectors", connName, "eventSubscriptions", name)
	return apiclient.HttpClient(u.String())
}

// DeleteEventSubscription
func DeleteEventSubscription(name string, connName string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	u.Path = path.Join(u.Path, "connectors", connName, "eventSubscriptions", name)
	return apiclient.HttpClient(u.String(), "", "DELETE")
}

// RetryEventSubscription
func RetryEventSubscription(name string, connName string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	u.Path = path.Join(u.Path, "connectors", connName, "eventSubscriptions", name+":retry")
	return apiclient.HttpClient(u.String(), "")
}

// ListEventSubscriptions
func ListEventSubscriptions(connName string, pageSize int, pageToken string, filter string, orderBy string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	u.Path = path.Join(u.Path, "connectors", connName, "eventSubscriptions")
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
