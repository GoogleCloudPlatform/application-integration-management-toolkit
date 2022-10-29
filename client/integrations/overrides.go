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
	TaskOverrides []taskconfig `json:"task_overrides,omitempty"`
}

// mergeOverrides
func mergeOverrides(eversion integrationVersionExternal, o overrides) integrationVersionExternal {

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
