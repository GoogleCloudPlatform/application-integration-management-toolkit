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

	"internal/apiclient"

	"internal/clilog"

	"internal/client/authconfigs"
	"internal/client/connections"
)

type overrides struct {
	TriggerOverrides    []triggeroverrides    `json:"trigger_overrides,omitempty"`
	TaskOverrides       []taskconfig          `json:"task_overrides,omitempty"`
	ConnectionOverrides []connectionoverrides `json:"connection_overrides,omitempty"`
	ParamOverrides      []parameterExternal   `json:"param_overrides,omitempty"`
}

type triggeroverrides struct {
	TriggerNumber                string  `json:"triggerNumber,omitempty"`
	TriggerType                  string  `json:"triggerType,omitempty"`
	ProjectId                    *string `json:"projectId,omitempty"`
	TopicName                    *string `json:"topicName,omitempty"`
	APIPath                      *string `json:"apiPath,omitempty"`
	CloudSchedulerServiceAccount *string `json:"cloudSchedulerServiceAccount,omitempty"`
	CloudSchedulerLocation       *string `json:"cloudSchedulerLocation,omitempty"`
	CloudSchedulerCronTab        *string `json:"cloudSchedulerCronTab,omitempty"`
}

type connectionoverrides struct {
	TaskId     string                   `json:"taskId,omitempty"`
	Task       string                   `json:"task,omitempty"`
	Parameters connectionoverrideparams `json:"parameters,omitempty"`
}

type connectionoverrideparams struct {
	ConnectionName     string `json:"connectionName,omitempty"`
	ConnectionLocation string `json:"connectionLocation,omitempty"`
}

type connectiondetails struct {
	Type       string           `json:"@type,omitempty"`
	Connection connectionparams `json:"connection,omitempty"`
	Operation  string           `json:"operation,omitempty"`
}

type connectionparams struct {
	ServiceName      string `json:"serviceName,omitempty"`
	ConnectionName   string `json:"connectionName,omitempty"`
	ConnectorVersion string `json:"connectorVersion,omitempty"`
}

const pubsubTrigger = "cloud_pubsub_external_trigger/projects/cloud-crm-eventbus-cpsexternal/subscriptions/"
const apiTrigger = "api_trigger/"

const authConfigValue = "{  \"@type\": \"type.googleapis.com/enterprise.crm.eventbus.authconfig.AuthConfigTaskParam\",\"authConfigId\": \""

// mergeOverrides
func mergeOverrides(eversion integrationVersionExternal, o overrides, suppressWarnings bool) (integrationVersionExternal, error) {
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
					if !suppressWarnings {
						clilog.Warning.Printf("unsupported trigger type %s\n", trigger.TriggerType)
					}
				}
				eversion.TriggerConfigs[triggerIndex] = trigger
				foundOverride = true
			}
		}
		if !foundOverride && !suppressWarnings {
			clilog.Warning.Printf("trigger override id %s was not found in the integration json\n",
				triggerOverride.TriggerNumber)
		}
	}

	//apply task overrides
	for _, taskOverride := range o.TaskOverrides {
		foundOverride := false
		for taskIndex, task := range eversion.TaskConfigs {
			if taskOverride.TaskId == task.TaskId && taskOverride.Task == task.Task && task.Task != "GenericConnectorTask" {
				task.Parameters = overrideParameters(taskOverride.Parameters, task.Parameters, suppressWarnings)
				eversion.TaskConfigs[taskIndex] = task
				foundOverride = true
			}
		}
		if !foundOverride && !suppressWarnings {
			clilog.Warning.Printf("task override %s with id %s was not found in the integration json\n",
				taskOverride.DisplayName, taskOverride.TaskId)
		}
	}

	for _, paramOverride := range o.ParamOverrides {
		foundOverride := false
		for ipIndex, ip := range eversion.IntegrationParameters {
			if paramOverride.Key == ip.Key {
				ip.DefaultValue = paramOverride.DefaultValue
			}
			eversion.IntegrationParameters[ipIndex] = ip
			foundOverride = true

		}
		if !foundOverride && !suppressWarnings {
			clilog.Warning.Printf("param override key %s with dataTpe %s was not found in the integration json\n",
				paramOverride.Key, paramOverride.DataType)
		}
	}

	//apply connection overrides
	if !apiclient.DryRun() {
		for _, connectionOverride := range o.ConnectionOverrides {
			foundOverride := false
			for taskIndex, task := range eversion.TaskConfigs {
				if connectionOverride.TaskId == task.TaskId && connectionOverride.Task == task.Task {
					newcp, err := getNewConnectionParams(connectionOverride.Parameters.ConnectionName,
						connectionOverride.Parameters.ConnectionLocation)
					if err != nil {
						return eversion, err
					}

					cparams := task.Parameters["config"]

					cd, err := getConnectionDetails(*cparams.Value.JsonValue)
					if err != nil {
						return eversion, err
					}

					cd.Connection.ConnectionName = newcp.ConnectionName
					cd.Connection.ConnectorVersion = newcp.ConnectorVersion
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
			if !foundOverride && !suppressWarnings {
				clilog.Warning.Printf("task override with id %s was not found in the integration json\n",
					connectionOverride.TaskId)
			}
		}
	}
	return eversion, nil
}

func extractOverrides(iversion integrationVersion) (overrides, error) {
	taskOverrides := overrides{}

	for _, task := range iversion.TaskConfigs {
		if task.Task == "GenericConnectorTask" {
			if err := handleGenericConnectorTask(task, &taskOverrides); err != nil {
				return taskOverrides, err
			}
		} else if task.Task == "GenericRestV2Task" {
			if err := handleGenericRestV2Task(task, &taskOverrides); err != nil {
				return taskOverrides, err
			}
		} else if task.Task == "CloudFunctionTask" {
			if err := handleCloudFunctionTask(task, &taskOverrides); err != nil {
				return taskOverrides, err
			}
		}
	}
	for _, param := range iversion.IntegrationParameters {
		if strings.HasPrefix(param.Key, "_") && !inputOutputVariable(param.InputOutputType) {
			ip := parameterExternal{}
			ip.Key = param.Key
			ip.DefaultValue = param.DefaultValue
			taskOverrides.ParamOverrides = append(taskOverrides.ParamOverrides, ip)
		}
	}
	for _, triggerConfig := range iversion.TriggerConfigs {
		if triggerConfig.TriggerType == "CLOUD_PUBSUB_EXTERNAL" {
			subscription := triggerConfig.Properties["Subscription name"]
			triggerOverride := triggeroverrides{}
			triggerOverride.ProjectId = new(string)
			triggerOverride.TopicName = new(string)
			*triggerOverride.ProjectId = strings.Split(subscription, "_")[0]
			*triggerOverride.TopicName = strings.Split(subscription, "_")[1]
			triggerOverride.TriggerNumber = triggerConfig.TriggerNumber
			taskOverrides.TriggerOverrides = append(taskOverrides.TriggerOverrides, triggerOverride)
		}
	}

	return taskOverrides, nil
}

func inputOutputVariable(variable string) bool {
	if variable == "IN" || variable == "OUT" || variable == "IN_OUT" {
		return true
	}
	return false
}

func handleGenericRestV2Task(taskConfig taskconfig, taskOverrides *overrides) error {
	tc := taskconfig{}
	tc.TaskId = taskConfig.TaskId
	tc.Task = taskConfig.Task
	tc.Parameters = map[string]eventparameter{}
	tc.Parameters["url"] = taskConfig.Parameters["url"]
	if _, ok := taskConfig.Parameters["authConfig"]; ok {
		displayName, err := authconfigs.GetDisplayName(getAuthConfigUuid(*taskConfig.Parameters["authConfig"].Value.JsonValue))
		if err != nil {
			return err
		}

		eventparam := eventparameter{}
		eventparam.Key = taskConfig.Parameters["authConfig"].Key
		eventparam.Value.StringValue = &displayName

		tc.Parameters["authConfig"] = eventparam
	}
	taskOverrides.TaskOverrides = append(taskOverrides.TaskOverrides, tc)
	return nil
}

func handleCloudFunctionTask(taskConfig taskconfig, taskOverrides *overrides) error {
	tc := taskconfig{}
	tc.TaskId = taskConfig.TaskId
	tc.Task = taskConfig.Task
	tc.Parameters = map[string]eventparameter{}
	tc.Parameters["TriggerUrl"] = taskConfig.Parameters["TriggerUrl"]
	if _, ok := taskConfig.Parameters["authConfig"]; ok {
		displayName, err := authconfigs.GetDisplayName(getAuthConfigUuid(*taskConfig.Parameters["authConfig"].Value.JsonValue))
		if err != nil {
			return err
		}

		eventparam := eventparameter{}
		eventparam.Key = taskConfig.Parameters["authConfig"].Key
		eventparam.Value.StringValue = &displayName

		tc.Parameters["authConfig"] = eventparam
	}
	taskOverrides.TaskOverrides = append(taskOverrides.TaskOverrides, tc)
	return nil
}

func handleGenericConnectorTask(taskConfig taskconfig, taskOverrides *overrides) error {
	co := connectionoverrides{}
	co.TaskId = taskConfig.TaskId
	co.Task = taskConfig.Task

	cparams, ok := taskConfig.Parameters["config"]
	if !ok {
		return nil
	}
	cd, err := getConnectionDetails(*cparams.Value.JsonValue)
	if err != nil {
		return err
	}

	parts := strings.Split(cd.Connection.ConnectionName, "/")
	connName := parts[len(parts)-1]
	co.Parameters.ConnectionName = connName

	taskOverrides.ConnectionOverrides = append(taskOverrides.ConnectionOverrides, co)
	return nil
}

// overrideParameters
func overrideParameters(overrideParameters map[string]eventparameter, taskParameters map[string]eventparameter,
	suppressWarnings bool) map[string]eventparameter {
	for overrideParamName, overrideParam := range overrideParameters {
		if overrideParam.Key == "authConfig" {
			apiclient.SetClientPrintHttpResponse(false)
			acversion, err := authconfigs.Find(*overrideParam.Value.StringValue, "")
			apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())
			if err != nil {
				clilog.Warning.Println(err)
				return taskParameters
			}
			*taskParameters[overrideParamName].Value.JsonValue = fmt.Sprintf("%s%s\"}", authConfigValue, acversion)
		} else {
			_, found := taskParameters[overrideParamName]
			if found {
				taskParameters[overrideParamName] = overrideParam
			} else {
				if !suppressWarnings {
					clilog.Warning.Printf("override param %s was not found\n", overrideParamName)
				}
			}
		}
	}
	return taskParameters
}

func getNewConnectionParams(connectionName string, connectionLocation string) (cp connectionparams, err error) {
	cp = connectionparams{}
	var connectionVersionResponse map[string]interface{}
	var integrationRegion string

	apiclient.SetClientPrintHttpResponse(false)
	defer apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())

	if connectionLocation != "" {
		integrationRegion = apiclient.GetRegion()     //store the integration location
		err = apiclient.SetRegion(connectionLocation) //set the connector region
		if err != nil {
			return cp, err
		}
	}
	connResp, err := connections.Get(connectionName, "BASIC", false, false) //get connector details
	if connectionLocation != "" {
		err = apiclient.SetRegion(integrationRegion) //set the integration region back
		if err != nil {
			return cp, err
		}
	}
	if err != nil {
		return cp, err
	}

	err = json.Unmarshal(connResp, &connectionVersionResponse)
	if err != nil {
		return cp, err
	}

	cp.ConnectorVersion = fmt.Sprintf("%v", connectionVersionResponse["connectorVersion"])
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
