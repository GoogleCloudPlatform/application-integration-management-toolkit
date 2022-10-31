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
	"strings"

	"github.com/apigee/apigeecli/clilog"
	"github.com/srinandan/integrationcli/apiclient"
	"github.com/srinandan/integrationcli/client/authconfigs"
	"github.com/srinandan/integrationcli/client/connections"
)

type overrides struct {
	TriggerOverrides    []triggeroverrides    `json:"trigger_overrides,omitempty"`
	TaskOverrides       []taskconfig          `json:"task_overrides,omitempty"`
	ConnectionOverrides []connectionoverrides `json:"connection_overrides,omitempty"`
}

type triggeroverrides struct {
	TriggerNumber string  `json:"triggerNumber,omitempty"`
	TriggerType   string  `json:"triggerType,omitempty"`
	ProjectId     *string `json:"projectId,omitempty"`
	TopicName     *string `json:"topicName,omitempty"`
	APIPath       *string `json:"apiPath,omitempty"`
}

type connectionoverrides struct {
	TaskId     string                   `json:"taskId,omitempty"`
	Task       string                   `json:"task,omitempty"`
	Parameters connectionoverrideparams `json:"parameters,omitempty"`
}

type connectionoverrideparams struct {
	ConnectionName string `json:"connectionName,omitempty"`
}

type connectiondetails struct {
	Type       string           `json:"@type,omitempty"`
	Connection connectionparams `json:"connection,omitempty"`
	Operation  string           `json:"operation,omitempty"`
}

type connectionparams struct {
	ServiceName       string `json:"serviceName,omitempty"`
	ConnectionName    string `json:"connectionName,omitempty"`
	ConnectionVersion string `json:"connectionVersion,omitempty"`
}

const pubsubTrigger = "cloud_pubsub_external_trigger/projects/cloud-crm-eventbus-cpsexternal/subscriptions/"
const apiTrigger = "api_trigger/"
const authConfigValue = "{  \"@type\": \"type.googleapis.com/enterprise.crm.eventbus.authconfig.AuthConfigTaskParam\",\"authConfigId\": \""

// mergeOverrides
func mergeOverrides(eversion integrationVersionExternal, o overrides, supressWarnings bool) (integrationVersionExternal, error) {

	//apply trigger overrides
	for _, triggerOverride := range o.TriggerOverrides {
		foundOverride := false
		for triggerIndex, trigger := range eversion.TriggerConfigs {
			if triggerOverride.TriggerNumber == trigger.TriggerNumber {
				switch trigger.TriggerType {
				case "CLOUD_PUBSUB_EXTERNAL":
					trigger.TriggerId = pubsubTrigger + *triggerOverride.ProjectId + "_" + *triggerOverride.TopicName
					trigger.Properties["Subscription name"] = *triggerOverride.ProjectId + "_" + *triggerOverride.TopicName
				case "API":
					trigger.TriggerId = apiTrigger + *triggerOverride.APIPath
				default:
					if !supressWarnings {
						clilog.Warning.Printf("unsupported trigger type %s\n", trigger.TriggerType)
					}
				}
				eversion.TriggerConfigs[triggerIndex] = trigger
				foundOverride = true
			}
		}
		if !foundOverride && !supressWarnings {
			clilog.Warning.Printf("trigger override id %s was not found in the integration json\n", triggerOverride.TriggerNumber)
		}
	}

	//apply task overrides
	for _, taskOverride := range o.TaskOverrides {
		foundOverride := false
		for taskIndex, task := range eversion.TaskConfigs {
			if taskOverride.TaskId == task.TaskId && taskOverride.Task == task.Task && task.Task != "GenericConnectorTask" {
				if task.Task == "CloudFunctionTask" {
					task.Parameters = overrideCfParameters(taskOverride.Parameters, task.Parameters, supressWarnings)
				} else {
					task.Parameters = overrideParameters(taskOverride.Parameters, task.Parameters, supressWarnings)
				}
				eversion.TaskConfigs[taskIndex] = task
				foundOverride = true
			}
		}
		if !foundOverride && !supressWarnings {
			clilog.Warning.Printf("task override %s with id %s was not found in the integration json\n", taskOverride.DisplayName, taskOverride.TaskId)
		}
	}

	//apply connection overrides
	for _, connectionOverride := range o.ConnectionOverrides {
		foundOverride := false
		for taskIndex, task := range eversion.TaskConfigs {
			if connectionOverride.TaskId == task.TaskId && connectionOverride.Task == task.Task {
				newcp, err := getNewConnectionParams(connectionOverride.Parameters.ConnectionName)
				if err != nil {
					return eversion, err
				}

				cparams := task.Parameters["config"]

				cd, err := getConnectionDetails(*cparams.Value.JsonValue)
				if err != nil {
					return eversion, err
				}

				cd.Connection.ConnectionName = newcp.ConnectionName
				cd.Connection.ConnectionVersion = newcp.ConnectionVersion
				cd.Connection.ServiceName = newcp.ServiceName

				jsonValue, err := stringifyValue(cd)
				if err != nil {
					return eversion, err
				}

				*cparams.Value.JsonValue = jsonValue
				task.Parameters["config"] = cparams
				eversion.TaskConfigs[taskIndex] = task

				foundOverride = true
			}
		}
		if !foundOverride && !supressWarnings {
			clilog.Warning.Printf("task override with id %s was not found in the integration json\n", connectionOverride.TaskId)
		}
	}
	return eversion, nil
}

// overrideParameters
func overrideParameters(overrideParameters map[string]eventparameter, taskParameters map[string]eventparameter, supressWarnings bool) map[string]eventparameter {
	for overrideParamName, overrideParam := range overrideParameters {
		_, found := taskParameters[overrideParamName]
		if found {
			taskParameters[overrideParamName] = overrideParam
		} else {
			if !supressWarnings {
				clilog.Warning.Printf("override param %s was not found\n", overrideParamName)
			}
		}
	}
	return taskParameters
}

// overrideCfParameters
func overrideCfParameters(overrideParameters map[string]eventparameter, taskParameters map[string]eventparameter, supessWarnings bool) map[string]eventparameter {
	for overrideParamName, overrideParam := range overrideParameters {
		if overrideParam.Key == "authConfig" {
			apiclient.SetPrintOutput(false)
			acversion, err := authconfigs.Find(*overrideParam.Value.StringValue, "")
			apiclient.SetPrintOutput(true)
			if err != nil {
				clilog.Warning.Println(err)
				return taskParameters
			}
			*taskParameters[overrideParamName].Value.JsonValue = fmt.Sprintf("%s\"}", acversion)
		} else {
			_, found := taskParameters[overrideParamName]
			if found {
				taskParameters[overrideParamName] = overrideParam
			} else {
				if !supessWarnings {
					clilog.Warning.Printf("override param %s was not found\n", overrideParamName)
				}
			}
		}
	}
	return taskParameters
}

func getNewConnectionParams(connectionName string) (connectionparams, error) {
	cp := connectionparams{}
	var connectionVersionResponse map[string]interface{}
	apiclient.SetPrintOutput(false)
	connResp, err := connections.Get(connectionName, "BASIC")
	apiclient.SetPrintOutput(true)
	if err != nil {
		return cp, err
	}

	err = json.Unmarshal(connResp, &connectionVersionResponse)
	if err != nil {
		return cp, err
	}

	cp.ConnectionVersion = fmt.Sprintf("%v", connectionVersionResponse["connectorVersion"])
	cp.ServiceName = fmt.Sprintf("%v", connectionVersionResponse["serviceDirectory"])
	cp.ConnectionName = fmt.Sprintf("%v", connectionVersionResponse["name"])

	return cp, nil
}

func getConnectionDetails(jsonValue string) (connectiondetails, error) {
	cd := connectiondetails{}
	t := strings.ReplaceAll(jsonValue, "\n", "")
	t = strings.ReplaceAll(t, "\\", "")
	err := json.Unmarshal([]byte(t), &cd)
	return cd, err
}

func stringifyValue(cd connectiondetails) (string, error) {
	jsonValue, err := json.Marshal(cd)
	if err != nil {
		return "", err
	}
	return string(jsonValue), nil
}
