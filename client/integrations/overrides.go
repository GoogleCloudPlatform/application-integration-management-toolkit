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
	"github.com/apigee/apigeecli/clilog"
)

type overrides struct {
	TriggerOverrides []triggeroverrides `json:"trigger_overrides,omitempty"`
	TaskOverrides    []taskconfig       `json:"task_overrides,omitempty"`
}

type triggeroverrides struct {
	TriggerNumber string  `json:"triggerNumber,omitempty"`
	TriggerType   string  `json:"triggerType,omitempty"`
	ProjectId     *string `json:"projectId,omitempty"`
	TopicName     *string `json:"topicName,omitempty"`
	APIPath       *string `json:"apiPath,omitempty"`
}

const pubsubTrigger = "cloud_pubsub_external_trigger/projects/cloud-crm-eventbus-cpsexternal/subscriptions/"
const apiTrigger = "api_trigger/"

// mergeOverrides
func mergeOverrides(eversion integrationVersionExternal, o overrides) integrationVersionExternal {

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
				}
				eversion.TriggerConfigs[triggerIndex] = trigger
				foundOverride = true
			}
		}
		if !foundOverride {
			clilog.Warning.Printf("trigger override id %s was not found in the integration json\n", triggerOverride.TriggerNumber)
		}
	}

	//apply task overrides
	for _, taskOverride := range o.TaskOverrides {
		foundOverride := false
		for taskIndex, task := range eversion.TaskConfigs {
			if taskOverride.TaskId == task.TaskId {
				task.Parameters = overrideParameters(taskOverride.Parameters, task.Parameters)
				eversion.TaskConfigs[taskIndex] = task
				foundOverride = true
			}
		}
		if !foundOverride {
			clilog.Warning.Printf("task override %s with id %s was not found in the integration json\n", taskOverride.DisplayName, taskOverride.TaskId)
		}
	}
	return eversion
}

// overrideParameters
func overrideParameters(overrideParameters map[string]eventparameter, taskParameters map[string]eventparameter) map[string]eventparameter {
	for overrideParamName, overrideParam := range overrideParameters {
		_, found := taskParameters[overrideParamName]
		if found {
			taskParameters[overrideParamName] = overrideParam
		} else {
			clilog.Warning.Printf("override param %s was not found\n", overrideParamName)
		}
	}
	return taskParameters
}
