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
	"fmt"
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
func ListExecutions(name string, pageSize int, pageToken string, filter string, orderBy string) (respBody []byte, err error) {
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
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

// Execute
func Execute(name string, content []byte) (respBody []byte, err error) {
	e := execute{}
	if err = json.Unmarshal(content, &e); err != nil {
		return nil, err
	}

	regExTrigger := regexp.MustCompile(`api_trigger\/\w+`)
	if !regExTrigger.MatchString(e.TriggerId) {
		return nil, fmt.Errorf("triggerId must match the format api_trigger/*")
	}

	apiclient.ClientPrintHttpResponse.Set(false)
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name+":execute")
	respBody, err = apiclient.HttpClient(u.String(), string(content))
	if err != nil {
		return nil, err
	}

	eresp := executionResponse{}
	err = json.Unmarshal(respBody, &eresp)
	if err != nil {
		return nil, err
	}

	// remove from response
	eresp.EventParameters = nil

	respBody, err = json.Marshal(eresp)
	if err != nil {
		return nil, err
	}

	apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
	apiclient.PrettyPrint(respBody)

	return respBody, err
}

func Cancel(name string, executionID string, cancelReason string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "executions", executionID, ":cancel")
	payload := "{ \"cancelReason\":\"" + cancelReason + "\"}"
	respBody, err = apiclient.HttpClient(u.String(), payload)
	return respBody, err
}

func Replay(name string, executionID string, replayReason string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "executions", executionID, ":replay")
	payload := "{ \"replayReason\":\"" + replayReason + "\"}"
	respBody, err = apiclient.HttpClient(u.String(), payload)
	return respBody, err
}
