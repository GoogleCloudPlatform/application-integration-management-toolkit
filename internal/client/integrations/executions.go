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
	"encoding/json"
	"internal/apiclient"
	"net/url"
	"path"
	"regexp"
	"strconv"
)

type execute struct {
	TriggerId           string                    `json:"triggerId,omitempty"`
	DoNotPropagateError bool                      `json:"doNotPropagateError,omitempty"`
	RequestId           string                    `json:"requestId,omitempty"`
	InputParameters     map[string]inputparameter `json:"inputParameters,omitempty"`
}

type inputparameter struct {
	StringValue  *string       `json:"stringValue,omitempty"`
	IntValue     *string       `json:"intValue,omitempty"`
	DoubleValue  *float32      `json:"doubleValue,omitempty"`
	BooleanValue *bool         `json:"booleanValue,omitempty"`
	JsonValue    *string       `json:"jsonValue,omitempty"`
	StringArray  *stringarray  `json:"stringArray,omitempty"`
	IntArray     *intarray     `json:"intArray,omitempty"`
	DoubleArray  *doublearray  `json:"doubleArray,omitempty"`
	BooleanArray *booleanarray `json:"booleanArray,omitempty"`
}

type stringarray struct {
	StringValues []string `json:"stringValues,omitempty"`
}

type intarray struct {
	IntValues []string `json:"intValues,omitempty"`
}

type doublearray struct {
	DoubleValues []float32 `json:"doubleValues,omitempty"`
}

type booleanarray struct {
	BooleanValues []bool
}

type executionResponse struct {
	ExecutionId      string              `json:"executionId,omitempty"`
	EventParameters  *parametersInternal `json:"eventParameters,omitempty"`
	OutputParameters interface{}         `json:"outputParameters,omitempty"`
}

// ListExecutions lists all executions
func ListExecutions(name string, pageSize int, pageToken string, filter string, orderBy string) apiclient.APIResponse {
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
	u.Path = path.Join(u.Path, "integrations", name, "executions")
	return apiclient.HttpClient(u.String())
}

// Execute
func Execute(name string, content []byte) apiclient.APIResponse {
	e := execute{}
	var err error
	if err = json.Unmarshal(content, &e); err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}

	regExTrigger := regexp.MustCompile(`api_trigger\/\w+`)
	if !regExTrigger.MatchString(e.TriggerId) {
		return apiclient.APIResponse{
			Err: apiclient.NewCliError("triggerId must match the format api_trigger/*", nil),
		}
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name+":execute")
	response := apiclient.HttpClient(u.String(), string(content))
	if response.Err != nil {
		return response
	}

	eresp := executionResponse{}
	err = json.Unmarshal(response.RespBody, &eresp)
	if err != nil {
		return apiclient.APIResponse{
			Err: apiclient.NewCliError("error unmarshalling", err),
		}
	}

	// remove from response
	eresp.EventParameters = nil

	response.RespBody, response.Err = json.Marshal(eresp)
	return response
}

func Cancel(name string, executionID string, cancelReason string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "executions", executionID, ":cancel")
	payload := "{ \"cancelReason\":\"" + cancelReason + "\"}"
	return apiclient.HttpClient(u.String(), payload)
}

func Replay(name string, executionID string, replayReason string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "executions", executionID, ":replay")
	payload := "{ \"replayReason\":\"" + replayReason + "\"}"
	return apiclient.HttpClient(u.String(), payload)
}
