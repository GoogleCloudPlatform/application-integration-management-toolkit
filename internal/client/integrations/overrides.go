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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"internal/apiclient"

	"internal/clilog"

	"internal/client/authconfigs"
	"internal/client/connections"
)

type overrides struct {
	TriggerOverrides     []triggeroverrides    `json:"trigger_overrides,omitempty"`
	TaskOverrides        []taskconfig          `json:"task_overrides,omitempty"`
	ConnectionOverrides  []connectionoverrides `json:"connection_overrides,omitempty"`
	ParamOverrides       []parameterExternal   `json:"param_overrides,omitempty"`
	IntegrationOverrides integrationoverrides  `json:"integration_overrides,omitempty"`
}

type integrationoverrides struct {
	RunAsServiceAccount       *string             `json:"runAsServiceAccount,omitempty"`
	DatabasePersistencePolicy string              `json:"databasePersistencePolicy"`
	EnableVariableMasking     bool                `json:"enableVariableMasking"`
	CloudLoggingDetails       cloudLoggingDetails `json:"cloudLoggingDetails,omitempty"`
}

type triggeroverrides struct {
	TriggerNumber                string            `json:"triggerNumber,omitempty"`
	TriggerType                  string            `json:"triggerType,omitempty"`
	ProjectId                    *string           `json:"projectId,omitempty"`
	TopicName                    *string           `json:"topicName,omitempty"`
	APIPath                      *string           `json:"apiPath,omitempty"`
	ServiceAccount               *string           `json:"serviceAccount,omitempty"`
	Properties                   map[string]string `json:"properties,omitempty"`
	CloudSchedulerServiceAccount *string           `json:"cloudSchedulerServiceAccount,omitempty"`
	CloudSchedulerLocation       *string           `json:"cloudSchedulerLocation,omitempty"`
	CloudSchedulerCronTab        *string           `json:"cloudSchedulerCronTab,omitempty"`
}

type connectionoverrides struct {
	TaskId     string                   `json:"taskId,omitempty"`
	Task       string                   `json:"task,omitempty"`
	Parameters connectionoverrideparams `json:"parameters,omitempty"`
}

type connectionoverrideparams struct {
	ConnectionName     string          `json:"connectionName,omitempty"`
	ConnectionLocation string          `json:"connectionLocation,omitempty"`
	EntityType         *eventparameter `json:"entityType,omitempty"`
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

const (
	pubsubTrigger = "cloud_pubsub_external_trigger/projects/cloud-crm-eventbus-cpsexternal/subscriptions/"
	apiTrigger    = "api_trigger/"
)

const authConfigValue = "{  \"@type\": \"type.googleapis.com/enterprise.crm.eventbus.authconfig.AuthConfigTaskParam\",\"authConfigId\": \""

// mergeOverrides
func mergeOverrides(eversion integrationVersionExternal, o overrides, grantPermission bool) (integrationVersionExternal, error) {
	// apply trigger overrides
	for _, triggerOverride := range o.TriggerOverrides {
		foundOverride := false
		for triggerIndex, trigger := range eversion.TriggerConfigs {
			if triggerOverride.TriggerNumber == trigger.TriggerNumber {
				switch trigger.TriggerType {
				case "CLOUD_PUBSUB_EXTERNAL":
					if triggerOverride.ProjectId == nil || triggerOverride.TopicName == nil {
						return eversion, fmt.Errorf("projectid and topicName are mandatory in the overrides")
					}
					trigger.TriggerId = pubsubTrigger + *triggerOverride.ProjectId + "_" + *triggerOverride.TopicName
					trigger.Properties["Subscription name"] = *triggerOverride.ProjectId + "_" + *triggerOverride.TopicName
					if triggerOverride.ServiceAccount != nil {
						serviceAccountName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", *triggerOverride.ServiceAccount, *triggerOverride.ProjectId)
						trigger.Properties["Service account"] = serviceAccountName
						if grantPermission {
							// create the SA if it doesn't exist
							if err := apiclient.CreateServiceAccount(serviceAccountName); err != nil {
								return eversion, err
							}
							if err := apiclient.SetIntegrationInvokerPermission(*triggerOverride.ProjectId, serviceAccountName); err != nil {
								clilog.Warning.Printf("Unable to update permissions for the service account: %v\n", err)
							}
						}
					}
				case "API":
					if triggerOverride.APIPath == nil {
						return eversion, fmt.Errorf("the field apiPath is missing from the API Trigger in overrides")
					}
					trigger.TriggerId = apiTrigger + *triggerOverride.APIPath
					if len(triggerOverride.Properties) > 0 {
						trigger.Properties = triggerOverride.Properties
					}
				case "CLOUD_SCHEDULER":
					if triggerOverride.CloudSchedulerServiceAccount != nil {
						trigger.CloudSchedulerConfig.ServiceAccountEmail = *triggerOverride.CloudSchedulerServiceAccount
					}
					if triggerOverride.CloudSchedulerCronTab != nil {
						trigger.CloudSchedulerConfig.CronTab = *triggerOverride.CloudSchedulerCronTab
					}
					if triggerOverride.CloudSchedulerLocation != nil {
						trigger.CloudSchedulerConfig.Location = *triggerOverride.CloudSchedulerLocation
					}
				case "INTEGRATION_CONNECTOR_TRIGGER":
					if len(triggerOverride.Properties) > 0 {
						trigger.TriggerId = fmt.Sprintf("integration_connector_trigger/projects/%s/locations/%s/connections/%s/eventSubscriptions/%s",
							triggerOverride.Properties["Project name"],
							triggerOverride.Properties["Region"], triggerOverride.Properties["Connection name"],
							triggerOverride.Properties["Subscription name"])
						trigger.Properties = triggerOverride.Properties
					}
				default:
					clilog.Warning.Printf("unsupported trigger type %s\n", trigger.TriggerType)
				}
				eversion.TriggerConfigs[triggerIndex] = trigger
				foundOverride = true
			}
		}
		if !foundOverride {
			clilog.Warning.Printf("trigger override id %s was not found in the integration json\n",
				triggerOverride.TriggerNumber)
		}
	}

	// apply task overrides
	for _, taskOverride := range o.TaskOverrides {
		foundOverride := false
		for taskIndex, task := range eversion.TaskConfigs {
			if taskOverride.TaskId == task.TaskId && taskOverride.Task == task.Task && task.Task != "GenericConnectorTask" {
				task.Parameters = overrideParameters(taskOverride.Parameters, task.Parameters)
				eversion.TaskConfigs[taskIndex] = task
				foundOverride = true
			}
		}
		if !foundOverride {
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
		if !foundOverride {
			clilog.Warning.Printf("param override key %s with dataTpe %s was not found in the integration json\n",
				paramOverride.Key, paramOverride.DataType)
		}
	}

	// apply connection overrides
	if !apiclient.DryRun() {
		foundOverride := false
		for _, connectionOverride := range o.ConnectionOverrides {
			for taskIndex, task := range eversion.TaskConfigs {
				if connectionOverride.TaskId == task.TaskId && connectionOverride.Task == task.Task {
					newcp, err := getNewConnectionParams(connectionOverride.Parameters.ConnectionName,
						connectionOverride.Parameters.ConnectionLocation)
					if err != nil {
						return eversion, err
					}

					cparams := task.Parameters["config"]
					// Google built connector
					if cparams.Value.JsonValue != nil {
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

						if connectionOverride.Parameters.EntityType != nil {
							task.Parameters["entityType"] = *connectionOverride.Parameters.EntityType
						}

						eversion.TaskConfigs[taskIndex] = task

						foundOverride = true
					}
					cversion := task.Parameters["connectionVersion"]
					// custom connector
					if cversion.Value.StringValue != nil {
						newcp, err := getNewConnectionParams(connectionOverride.Parameters.ConnectionName,
							connectionOverride.Parameters.ConnectionLocation)
						if err != nil {
							return eversion, err
						}
						*task.Parameters["connectionVersion"].Value.StringValue = newcp.ConnectorVersion
						*task.Parameters["connectionName"].Value.StringValue = newcp.ConnectionName
						eversion.TaskConfigs[taskIndex] = task
						foundOverride = true
					}
				}
			}
			if !foundOverride {
				clilog.Warning.Printf("task override with id %s was not found in the integration json\n",
					connectionOverride.TaskId)
			}
		}
	}

	// apply integration overrides

	if o.IntegrationOverrides.DatabasePersistencePolicy != "" {
		eversion.DatabasePersistencePolicy = o.IntegrationOverrides.DatabasePersistencePolicy
	}

	eversion.CloudLoggingDetails.CloudLoggingSeverity = o.IntegrationOverrides.CloudLoggingDetails.CloudLoggingSeverity
	eversion.CloudLoggingDetails.EnableCloudLogging = o.IntegrationOverrides.CloudLoggingDetails.EnableCloudLogging

	if o.IntegrationOverrides.RunAsServiceAccount != nil {
		eversion.RunAsServiceAccount = *o.IntegrationOverrides.RunAsServiceAccount
	}

	eversion.EnableVariableMasking = o.IntegrationOverrides.EnableVariableMasking

	return eversion, nil
}

func extractOverrides(iversion integrationVersion) (overrides, error) {
	taskOverrides := overrides{
		IntegrationOverrides: integrationoverrides{
			RunAsServiceAccount:       nil,
			DatabasePersistencePolicy: "DATABASE_PERSISTENCE_POLICY_UNSPECIFIED",
			EnableVariableMasking:     false,
			CloudLoggingDetails: cloudLoggingDetails{
				EnableCloudLogging:   false,
				CloudLoggingSeverity: "CLOUD_LOGGING_SEVERITY_UNSPECIFIED",
			},
		},
	}

	for _, task := range iversion.TaskConfigs {
		if task.Task == "GenericConnectorTask" {
			if err := handleGenericConnectorTask(task, &taskOverrides, iversion.IntegrationConfigParameters); err != nil {
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
			if param.DefaultValue != nil {
				ip.DefaultValue = param.DefaultValue
			}
			taskOverrides.ParamOverrides = append(taskOverrides.ParamOverrides, ip)
		}
	}
	for _, triggerConfig := range iversion.TriggerConfigs {
		switch triggerConfig.TriggerType {
		case "CLOUD_PUBSUB_EXTERNAL":
			subscription := triggerConfig.Properties["Subscription name"]
			triggerOverride := triggeroverrides{}
			triggerOverride.ProjectId = new(string)
			triggerOverride.TopicName = new(string)
			triggerOverride.ServiceAccount = new(string)
			*triggerOverride.ProjectId = strings.Split(subscription, "_")[0]
			*triggerOverride.TopicName = strings.Split(subscription, "_")[1]
			triggerOverride.TriggerNumber = triggerConfig.TriggerNumber
			*triggerOverride.ServiceAccount = triggerConfig.Properties["Service account"]
			taskOverrides.TriggerOverrides = append(taskOverrides.TriggerOverrides, triggerOverride)
		case "CLOUD_SCHEDULER":
			triggerOverride := triggeroverrides{}
			triggerOverride.CloudSchedulerServiceAccount = new(string)
			triggerOverride.CloudSchedulerLocation = new(string)
			triggerOverride.CloudSchedulerCronTab = new(string)
			*triggerOverride.CloudSchedulerServiceAccount = triggerConfig.CloudSchedulerConfig.ServiceAccountEmail
			*triggerOverride.CloudSchedulerLocation = triggerConfig.CloudSchedulerConfig.Location
			*triggerOverride.CloudSchedulerCronTab = triggerConfig.CloudSchedulerConfig.CronTab
			taskOverrides.TriggerOverrides = append(taskOverrides.TriggerOverrides, triggerOverride)
		case "INTEGRATION_CONNECTOR_TRIGGER":
			triggerOverride := triggeroverrides{}
			triggerOverride.Properties = triggerConfig.Properties
			triggerOverride.TriggerNumber = triggerConfig.TriggerNumber
			triggerOverride.TriggerType = triggerConfig.TriggerType
			taskOverrides.TriggerOverrides = append(taskOverrides.TriggerOverrides, triggerOverride)
		}
	}

	// handle integration overrides

	if iversion.DatabasePersistencePolicy != "" {
		taskOverrides.IntegrationOverrides.DatabasePersistencePolicy = iversion.DatabasePersistencePolicy
	}

	if iversion.RunAsServiceAccount != "" {
		taskOverrides.IntegrationOverrides.RunAsServiceAccount = new(string)
		*taskOverrides.IntegrationOverrides.RunAsServiceAccount = iversion.RunAsServiceAccount
	}

	if iversion.EnableVariableMasking {
		taskOverrides.IntegrationOverrides.EnableVariableMasking = iversion.EnableVariableMasking
	}
	if iversion.CloudLoggingDetails.CloudLoggingSeverity != "" {
		taskOverrides.IntegrationOverrides.CloudLoggingDetails.CloudLoggingSeverity = iversion.CloudLoggingDetails.CloudLoggingSeverity
	}
	if iversion.CloudLoggingDetails.EnableCloudLogging {
		taskOverrides.IntegrationOverrides.CloudLoggingDetails.EnableCloudLogging = iversion.CloudLoggingDetails.EnableCloudLogging
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

	// store in overrides only if config variables are not used
	urlEventParam := taskConfig.Parameters["url"]
	if urlEventParam.Value.StringValue != nil && !strings.HasPrefix(*urlEventParam.Value.StringValue, "$`CONFIG_") {
		tc.Parameters["url"] = taskConfig.Parameters["url"]
	} else if urlEventParam.Value.IntValue != nil && !strings.HasPrefix(*urlEventParam.Value.IntValue, "$`CONFIG_") {
		tc.Parameters["url"] = taskConfig.Parameters["url"]
	}

	if _, ok := taskConfig.Parameters["authConfig"]; ok {
		displayName, err := authconfigs.GetDisplayName(getAuthConfigUuid(*taskConfig.Parameters["authConfig"].Value.JsonValue))
		if err != nil {
			return err
		}
		if displayName != "" {
			eventparam := eventparameter{}
			eventparam.Key = taskConfig.Parameters["authConfig"].Key
			eventparam.Value.StringValue = &displayName

			tc.Parameters["authConfig"] = eventparam

		}
	}

	if len(tc.Parameters) > 0 {
		taskOverrides.TaskOverrides = append(taskOverrides.TaskOverrides, tc)
	}
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
		if displayName != "" {
			eventparam := eventparameter{}
			eventparam.Key = taskConfig.Parameters["authConfig"].Key
			eventparam.Value.StringValue = &displayName

			tc.Parameters["authConfig"] = eventparam
		}
	}
	taskOverrides.TaskOverrides = append(taskOverrides.TaskOverrides, tc)
	return nil
}

func handleGenericConnectorTask(taskConfig taskconfig, taskOverrides *overrides, iconfigParam []parameterConfig) error {
	co := connectionoverrides{}
	co.TaskId = taskConfig.TaskId
	co.Task = taskConfig.Task

	cparams, ok := taskConfig.Parameters["config"]
	connectionNameparams, okConnectionName := taskConfig.Parameters["connectionName"]

	if !ok && !okConnectionName {
		return nil
	}
	if connectionNameparams.Key == "connectionName" {
		if connectionNameparams.Value.StringValue != nil {
			connectionName, err := getConnectionStringFromConnectionName(*connectionNameparams.Value.StringValue, iconfigParam)
			if err != nil {
				return err
			}
			parts := strings.Split(connectionName, "/")
			connName := parts[len(parts)-1]
			co.Parameters.ConnectionName = connName

		}
	} else if (eventparameter{}) != cparams && ok {
		if cparams.Value.JsonValue != nil {
			cd, err := getConnectionDetails(*cparams.Value.JsonValue)
			if err != nil {
				return err
			}

			parts := strings.Split(cd.Connection.ConnectionName, "/")
			connName := parts[len(parts)-1]
			co.Parameters.ConnectionName = connName
		}
	}
	taskOverrides.ConnectionOverrides = append(taskOverrides.ConnectionOverrides, co)

	return nil
}

// overrideParameters
func overrideParameters(overrideParameters map[string]eventparameter,
	taskParameters map[string]eventparameter,
) map[string]eventparameter {
	for overrideParamName, overrideParam := range overrideParameters {
		if overrideParam.Key == "authConfig" {
			apiclient.ClientPrintHttpResponse.Set(false)
			acversion, err := authconfigs.Find(*overrideParam.Value.StringValue, "")
			apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
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
				clilog.Warning.Printf("override param %s was not found\n", overrideParamName)
			}
		}
	}
	return taskParameters
}

func getNewConnectionParams(connectionName string, connectionLocation string) (cp connectionparams, err error) {
	cp = connectionparams{}
	var connectionVersionResponse map[string]interface{}
	var integrationRegion string

	apiclient.ClientPrintHttpResponse.Set(false)
	defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

	if connectionLocation != "" {
		integrationRegion = apiclient.GetRegion()     // store the integration location
		err = apiclient.SetRegion(connectionLocation) // set the connector region
		if err != nil {
			return cp, err
		}
	}
	connResp, err := connections.Get(connectionName, "BASIC", false, false) // get connector details
	if connectionLocation != "" {
		err = apiclient.SetRegion(integrationRegion) // set the integration region back
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

// getConnectionStringFromConnectionName
func getConnectionStringFromConnectionName(connectionName string, iconfigParam []parameterConfig) (connection string, err error) {
	var name string
	if strings.HasPrefix(connectionName, "$`CONFIG_") {
		for _, param := range iconfigParam {
			if param.Parameter.Key == strings.ReplaceAll(connectionName, "$", "") {
				if param.Value != nil {
					name = *param.Value.StringValue
				} else if param.Parameter.DefaultValue != nil {
					name = *param.Parameter.DefaultValue.StringValue
				}
			}
		}
	} else {
		name = connectionName
	}

	re := regexp.MustCompile(`projects/(.*)/locations/(.*)/connections/(.*)`)
	if !re.MatchString(name) {
		return "", errors.New("Connection Name is not valid. Connection name should be in the format: projects/{projectId}/locations/{locationId}/connections/{connectionId}")
	}
	return name, nil
}
