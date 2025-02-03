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
	"internal/apiclient"
	"internal/client/authconfigs"
	"internal/clilog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const maxPageSize = 1000

// integrationInfo contains information about an Integration Flow to export
type integrationInfo struct {
	Name string
	Path string
}

type uploadIntegrationFormat struct {
	Content    string `json:"content" binding:"required"`
	FileFormat string `json:"fileFormat"`
}

type listIntegrationVersions struct {
	IntegrationVersions []integrationVersion `json:"integrationVersions,omitempty"`
	NextPageToken       string               `json:"nextPageToken,omitempty"`
}

type integrationVersion struct {
	Name                          string                   `json:"name,omitempty"`
	Description                   string                   `json:"description,omitempty"`
	TaskConfigsInternal           []map[string]interface{} `json:"taskConfigsInternal,omitempty"`
	TriggerConfigsInternal        []map[string]interface{} `json:"triggerConfigsInternal,omitempty"`
	IntegrationParametersInternal parametersInternal       `json:"integrationParametersInternal,omitempty"`
	Origin                        string                   `json:"origin,omitempty"`
	Status                        string                   `json:"status,omitempty"`
	SnapshotNumber                string                   `json:"snapshotNumber,omitempty"`
	UpdateTime                    string                   `json:"updateTime,omitempty"`
	LockHolder                    string                   `json:"lockHolder,omitempty"`
	CreateTime                    string                   `json:"createTime,omitempty"`
	LastModifierEmail             string                   `json:"lastModifierEmail,omitempty"`
	State                         string                   `json:"state,omitempty"`
	TriggerConfigs                []triggerconfig          `json:"triggerConfigs,omitempty"`
	TaskConfigs                   []taskconfig             `json:"taskConfigs,omitempty"`
	IntegrationParameters         []parameterExternal      `json:"integrationParameters,omitempty"`
	IntegrationConfigParameters   []parameterConfig        `json:"integrationConfigParameters,omitempty"`
	UserLabel                     *string                  `json:"userLabel,omitempty"`
	DatabasePersistencePolicy     string                   `json:"databasePersistencePolicy,default=DATABASE_PERSISTENCE_POLICY_UNSPECIFIED"`
	ErrorCatcherConfigs           []errorCatcherConfig     `json:"errorCatcherConfigs,omitempty"`
	RunAsServiceAccount           string                   `json:"runAsServiceAccount,omitempty"`
	ParentTemplateId              string                   `json:"parentTemplateId,omitempty"`
	CloudLoggingDetails           cloudLoggingDetails      `json:"cloudLoggingDetails,omitempty"`
	EnableVariableMasking         bool                     `json:"enableVariableMasking,omitempty"`
}

type integrationVersionExternal struct {
	Description                 string               `json:"description,omitempty"`
	SnapshotNumber              string               `json:"snapshotNumber,omitempty"`
	TriggerConfigs              []triggerconfig      `json:"triggerConfigs,omitempty"`
	TaskConfigs                 []taskconfig         `json:"taskConfigs,omitempty"`
	IntegrationParameters       []parameterExternal  `json:"integrationParameters,omitempty"`
	IntegrationConfigParameters []parameterConfig    `json:"integrationConfigParameters,omitempty"`
	UserLabel                   *string              `json:"userLabel,omitempty"`
	DatabasePersistencePolicy   string               `json:"databasePersistencePolicy,default=DATABASE_PERSISTENCE_POLICY_UNSPECIFIED"`
	ErrorCatcherConfigs         []errorCatcherConfig `json:"errorCatcherConfigs,omitempty"`
	RunAsServiceAccount         string               `json:"runAsServiceAccount,omitempty"`
	ParentTemplateId            string               `json:"parentTemplateId,omitempty"`
	CloudLoggingDetails         cloudLoggingDetails  `json:"cloudLoggingDetails,omitempty"`
	EnableVariableMasking       bool                 `json:"enableVariableMasking,omitempty"`
}

type cloudLoggingDetails struct {
	CloudLoggingSeverity string `json:"cloudLoggingSeverity,default=CLOUD_LOGGING_SEVERITY_UNSPECIFIED"`
	EnableCloudLogging   bool   `json:"enableCloudLogging"`
}

type listbasicIntegrationVersions struct {
	BasicIntegrationVersions []basicIntegrationVersion `json:"integrationVersions,omitempty"`
	NextPageToken            string                    `json:"nextPageToken,omitempty"`
}

type basicIntegrationVersion struct {
	Version        string `json:"version,omitempty"`
	SnapshotNumber string `json:"snapshotNumber,omitempty"`
	State          string `json:"state,omitempty"`
}

type listintegrations struct {
	Integrations  []integration `json:"integrations,omitempty"`
	NextPageToken string        `json:"nextPageToken,omitempty"`
}

type integration struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	UpdateTime  string `json:"updateTime,omitempty"`
	Active      bool   `json:"active,omitempty"`
}

type parametersInternal struct {
	Parameters []parameterInternal `json:"parameters,omitempty"`
}

type parameterInternal struct {
	Key         string     `json:"key,omitempty"`
	DataType    string     `json:"dataType,omitempty"`
	Name        string     `json:"name,omitempty"`
	IsTransient bool       `json:"isTransient,omitempty"`
	ProducedBy  producedBy `json:"producedBy,omitempty"`
	Producer    string     `json:"producer,omitempty"`
	Masked      bool       `json:"masked,omitempty"`
}

type parameterExternal struct {
	Key             string     `json:"key,omitempty"`
	DataType        string     `json:"dataType,omitempty"`
	DefaultValue    *valueType `json:"defaultValue,omitempty"`
	Name            string     `json:"name,omitempty"`
	IsTransient     bool       `json:"isTransient,omitempty"`
	InputOutputType string     `json:"inputOutputType,omitempty"`
	Producer        string     `json:"producer,omitempty"`
	Searchable      bool       `json:"searchable,omitempty"`
	JsonSchema      string     `json:"jsonSchema,omitempty"`
	Masked          bool       `json:"masked,omitempty"`
}

type parameterConfig struct {
	Parameter parameter  `json:"parameter,omitempty"`
	Value     *valueType `json:"value,omitempty"`
}

type parameter struct {
	Key          string     `json:"key,omitempty"`
	DataType     string     `json:"dataType,omitempty"`
	DefaultValue *valueType `json:"defaultValue,omitempty"`
	DisplayName  string     `json:"displayName,omitempty"`
}

type producedBy struct {
	ElementType       string `json:"elementType,omitempty"`
	ElementIdentifier string `json:"elementIdentifier,omitempty"`
}

type triggerconfig struct {
	Label                    string                   `json:"label,omitempty"`
	TriggerType              string                   `json:"triggerType,omitempty"`
	TriggerNumber            string                   `json:"triggerNumber,omitempty"`
	TriggerId                string                   `json:"triggerId,omitempty"`
	Description              string                   `json:"description,omitempty"`
	StartTasks               []nextTask               `json:"startTasks,omitempty"`
	NextTasksExecutionPolicy string                   `json:"nextTasksExecutionPolicy,omitempty"`
	AlertConfig              []map[string]interface{} `json:"alterConfig,omitempty"`
	Properties               map[string]string        `json:"properties,omitempty"`
	CloudSchedulerConfig     *cloudSchedulerConfig    `json:"cloudSchedulerConfig,omitempty"`
	ErrorCatcherId           string                   `json:"errorCatcherId,omitempty"`
}

type taskconfig struct {
	Task                         string                    `json:"task,omitempty"`
	TaskId                       string                    `json:"taskId,omitempty"`
	Parameters                   map[string]eventparameter `json:"parameters,omitempty"`
	DisplayName                  string                    `json:"displayName,omitempty"`
	NextTasks                    []nextTask                `json:"nextTasks,omitempty"`
	NextTasksExecutionPolicy     string                    `json:"nextTasksExecutionPolicy,omitempty"`
	TaskExecutionStrategy        string                    `json:"taskExecutionStrategy,omitempty"`
	JsonValidationOption         string                    `json:"jsonValidationOption,omitempty"`
	SuccessPolicy                *successPolicy            `json:"successPolicy,omitempty"`
	TaskTemplate                 string                    `json:"taskTemplate,omitempty"`
	FailurePolicy                *failurePolicy            `json:"failurePolicy,omitempty"`
	ConditionalFailurePolicies   *conditionalFailurePolicy `json:"conditionalFailurePolicies,omitempty"`
	SynchronousCallFailurePolicy *failurePolicy            `json:"synchronousCallFailurePolicy,omitempty"`
	ErrorCatcherId               string                    `json:"errorCatcherId,omitempty"`
	ExternalTaskType             string                    `json:"externalTaskType,omitempty"`
}

type errorCatcherConfig struct {
	Label              string  `json:"label,omitempty"`
	ErrorCatcherNumber string  `json:"errorCatcherNumber,omitempty"`
	ErrorCatcherId     string  `json:"errorCatcherId,omitempty"`
	StartErrorTasks    []tasks `json:"startErrorTasks,omitempty"`
}

type tasks struct {
	TaskId string `json:"taskId,omitempty"`
}

type eventparameter struct {
	Key    string    `json:"key,omitempty"`
	Value  valueType `json:"value,omitempty"`
	Masked bool      `json:"masked,omitempty"`
}

type valueType struct {
	StringValue  *string          `json:"stringValue,omitempty"`
	IntValue     *string          `json:"intValue,omitempty"`
	BooleanValue *bool            `json:"booleanValue,omitempty"`
	StringArray  *stringarraytype `json:"stringArray,omitempty"`
	JsonValue    *string          `json:"jsonValue,omitempty"`
	DoubleValue  float64          `json:"doubleValue,omitempty"`
	IntArray     *intarray        `json:"intArray,omitempty"`
	DoubleArray  *doublearray     `json:"doubleArray,omitempty"`
	BooleanArray *booleanarray    `json:"booleanArray,omitempty"`
}

type stringarraytype struct {
	StringValues []string `json:"stringValues,omitempty"`
}

type nextTask struct {
	TaskConfigId string `json:"taskConfigId,omitempty"`
	TaskId       string `json:"taskId,omitempty"`
	Condition    string `json:"condition,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Description  string `json:"description,omitempty"`
}

type successPolicy struct {
	FinalState string `json:"finalState,omitempty"`
}

type failurePolicy struct {
	RetryStrategy string `json:"retryStrategy,omitempty"`
	MaxRetries    int    `json:"maxRetries,omitempty"`
	IntervalTime  string `json:"intervalTime,omitempty"`
	Condition     string `json:"condition,omitempty"`
}

type cloudSchedulerConfig struct {
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty"`
	CronTab             string `json:"cronTab,omitempty"`
	Location            string `json:"location,omitempty"`
	ErrorMessage        string `json:"errorMessage,omitempty"`
}

type integrationConnection struct {
	Name             string
	Region           string
	Version          string
	CustomConnection bool
}

type conditionalFailurePolicy struct {
	FailurePolicies      []failurePolicy `json:"failurePolicies,omitempty"`
	DefaultFailurePolicy *failurePolicy  `json:"defaultFailurePolicy,omitempty"`
}

// CreateVersion
func CreateVersion(name string, content []byte, overridesContent []byte, snapshot string,
	userlabel string, grantPermission bool, basicInfo bool,
) apiclient.APIResponse {

	var err error
	iversion := integrationVersion{}
	if err := json.Unmarshal(content, &iversion); err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	// remove any internal elements if exists
	eversion := convertInternalToExternal(iversion)

	// merge overrides if overrides were provided
	if len(overridesContent) > 0 {
		o := overrides{
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

		if err = json.Unmarshal(overridesContent, &o); err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      err,
			}
		}
		if eversion, err = mergeOverrides(eversion, o, grantPermission); err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      err,
			}
		}
	}

	if snapshot != "" {
		eversion.SnapshotNumber = snapshot
	}

	if userlabel != "" {
		eversion.UserLabel = new(string)
		*eversion.UserLabel = userlabel
	}

	if content, err = json.Marshal(eversion); err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions")
	response := apiclient.HttpClient(u.String(), string(content))
	if response.Err != nil {
		return response
	}

	if basicInfo {
		response = getBasicInfo(response.RespBody)
		return response
	}

	return response
}

// Upload
func Upload(name string, content []byte) apiclient.APIResponse {
	uploadVersion := uploadIntegrationFormat{}
	if err := json.Unmarshal(content, &uploadVersion); err != nil {
		clilog.Error.Println("invalid format for upload. Upload must have the json field content which contains " +
			"stringified integration json and optionally the file format")
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	if uploadVersion.Content == "" {
		return apiclient.APIResponse{
			RespBody: nil,
			Err: fmt.Errorf("invalid format for upload. Upload must have the json field content which contains " +
				"stringified integration json and optionally the file format"),
		}
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions:upload")
	return apiclient.HttpClient(u.String(), string(content))
}

// Patch
func Patch(name string, version string, content []byte) apiclient.APIResponse {
	iversion := integrationVersion{}
	var err error

	if err = json.Unmarshal(content, &iversion); err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	// remove any internal elements if exists
	eversion := convertInternalToExternal(iversion)

	if content, err = json.Marshal(eversion); err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)
	return apiclient.HttpClient(u.String(), string(content), "PATCH")
}

// TakeOverEditLock
func TakeoverEditLock(name string, version string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)
	return apiclient.HttpClient(u.String(), "")
}

// ListVersions
func ListVersions(name string, pageSize int, pageToken string, filter string, orderBy string,
	allVersions bool, download bool, basicInfo bool,
) apiclient.APIResponse {

	var err error
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

	u.Path = path.Join(u.Path, "integrations", name, "versions")

	if !allVersions {
		if basicInfo {
			response := apiclient.HttpClient(u.String())
			if response.Err != nil {
				return response
			}
			listIvers := listIntegrationVersions{}
			listBIvers := listbasicIntegrationVersions{}

			if err := json.Unmarshal(response.RespBody, &listIvers); err != nil {
				return apiclient.APIResponse{
					RespBody: nil,
					Err:      err,
				}
			}

			for _, iVer := range listIvers.IntegrationVersions {
				basicIVer := basicIntegrationVersion{}
				basicIVer.SnapshotNumber = iVer.SnapshotNumber
				basicIVer.Version = getVersion(iVer.Name)
				basicIVer.State = iVer.State
				listBIvers.BasicIntegrationVersions = append(listBIvers.BasicIntegrationVersions, basicIVer)
			}
			newResp, err := json.Marshal(listBIvers)
			return apiclient.APIResponse{
				RespBody: newResp,
				Err:      err,
			}
		}
		response := apiclient.HttpClient(u.String())
		return response
	} else {
		response := apiclient.HttpClient(u.String())
		if response.Err != nil {
			return response
		}

		iversions := listIntegrationVersions{}
		if err := json.Unmarshal(response.RespBody, &iversions); err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      err,
			}
		}

		if apiclient.GetExportToFile() != "" {
			// Write each version to a file
			for _, iversion := range iversions.IntegrationVersions {
				var iversionBytes []byte
				if iversionBytes, err = json.Marshal(iversion); err != nil {
					return apiclient.APIResponse{
						RespBody: nil,
						Err:      err,
					}
				}
				version := iversion.Name[strings.LastIndex(iversion.Name, "/")+1:]
				fileName := strings.Join([]string{name, iversion.SnapshotNumber, version}, "+") + ".json"
				if download {
					version := iversion.Name[strings.LastIndex(iversion.Name, "/")+1:]
					response := Download(name, version)
					if response.Err != nil {
						return response
					}
					if err = apiclient.WriteByteArrayToFile(
						path.Join(apiclient.GetExportToFile(), fileName),
						false,
						response.RespBody); err != nil {
						return apiclient.APIResponse{
							RespBody: nil,
							Err:      err,
						}
					}
				} else {
					if err = apiclient.WriteByteArrayToFile(
						path.Join(apiclient.GetExportToFile(), fileName),
						false,
						iversionBytes); err != nil {
						return apiclient.APIResponse{
							RespBody: nil,
							Err:      err,
						}
					}
				}
				clilog.Info.Printf("Downloaded version %s for Integration flow %s\n", version, name)
			}
		}

		// if more versions exist, repeat the process
		if iversions.NextPageToken != "" {
			if response = ListVersions(name, -1, iversions.NextPageToken, filter, orderBy, true, download, false); response.Err != nil {
				return response
			}
		} else {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      nil,
			}
		}
	}
	return apiclient.APIResponse{
		RespBody: nil,
		Err:      nil,
	}
}

// List
func List(pageSize int, pageToken string, filter string, orderBy string) apiclient.APIResponse {
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
	u.Path = path.Join(u.Path, "integrations")
	return apiclient.HttpClient(u.String())
}

// Get
func Get(name string, version string, basicInfo bool, minimal bool, override bool) apiclient.APIResponse {
	if (basicInfo && minimal) || (basicInfo && override) || (minimal && override) {
		return apiclient.APIResponse{
			Err: errors.New("cannot combine basicInfo, minimal and override flags"),
		}
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)

	response := apiclient.HttpClient(u.String())

	if !override && !minimal && !basicInfo {
		return response
	}

	iversion := integrationVersion{}
	err := json.Unmarshal(response.RespBody, &iversion)
	if err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	if basicInfo {
		return getBasicInfo(response.RespBody)
	}

	if minimal {
		eversion := convertInternalToExternal(iversion)
		response.RespBody, response.Err = json.Marshal(eversion)
		return response
	}

	if override {
		var or overrides
		if or, err = extractOverrides(iversion); err != nil {
			return apiclient.APIResponse{
				Err: err,
			}
		}
		response.RespBody, response.Err = json.Marshal(or)
		return response
	}
	return response
}

// GetBySnapshot
func GetBySnapshot(name string, snapshot string, basicInfo bool, minimal bool, override bool) apiclient.APIResponse {
	response := ListVersions(name, -1, "", "snapshotNumber="+snapshot, "", false, false, true)
	if response.Err != nil {
		return response
	}

	listBasicVersions := listbasicIntegrationVersions{}
	err := json.Unmarshal(response.RespBody, &listBasicVersions)
	if err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}

	if len(listBasicVersions.BasicIntegrationVersions) < 1 {
		return apiclient.APIResponse{
			Err: fmt.Errorf("snapshot number was not found"),
		}
	}

	version := getVersion(listBasicVersions.BasicIntegrationVersions[0].Version)
	return Get(name, version, basicInfo, minimal, override)
}

// GetByUserlabel
func GetByUserlabel(name string, userLabel string, basicInfo bool, minimal bool, override bool) apiclient.APIResponse {
	response := ListVersions(name, -1, "", "userLabel="+userLabel, "", false, false, true)
	if response.Err != nil {
		return response
	}

	listBasicVersions := listbasicIntegrationVersions{}
	err := json.Unmarshal(response.RespBody, &listBasicVersions)
	if err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}

	if len(listBasicVersions.BasicIntegrationVersions) < 1 {
		return apiclient.APIResponse{
			Err: fmt.Errorf("userLabel was not found"),
		}
	}

	version := getVersion(listBasicVersions.BasicIntegrationVersions[0].Version)
	return Get(name, version, false, minimal, override)
}

// GetConfigVariables
func GetConfigVariables(contents []byte) apiclient.APIResponse {
	iversion := integrationVersion{}
	configVariables := make(map[string]interface{})
	var response apiclient.APIResponse
	var err error

	err = json.Unmarshal(contents, &iversion)
	if err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}

	for _, param := range iversion.IntegrationConfigParameters {
		configVariables[param.Parameter.Key] = ""
		if param.Value != nil {
			if param.Value.StringValue != nil {
				configVariables[param.Parameter.Key] = param.Value.StringValue
			} else if param.Value.IntValue != nil {
				configVariables[param.Parameter.Key], _ = strconv.ParseInt(*param.Value.IntValue, 10, 0)
			} else if param.Value.JsonValue != nil {
				configVariables[param.Parameter.Key] = getJson(*param.Value.JsonValue)
			} else if param.Value.BooleanValue != nil {
				configVariables[param.Parameter.Key] = param.Value.BooleanValue
			} else if param.Value.StringArray != nil {
				configVariables[param.Parameter.Key] = param.Value.StringArray.StringValues
			}
		} else if param.Parameter.DefaultValue != nil {
			if param.Parameter.DefaultValue.StringValue != nil {
				configVariables[param.Parameter.Key] = param.Parameter.DefaultValue.StringValue
			} else if param.Parameter.DefaultValue.IntValue != nil {
				configVariables[param.Parameter.Key], _ = strconv.ParseInt(*param.Parameter.DefaultValue.IntValue, 10, 0)
			} else if param.Parameter.DefaultValue.JsonValue != nil {
				configVariables[param.Parameter.Key] = getJson(*param.Parameter.DefaultValue.JsonValue)
			} else if param.Parameter.DefaultValue.BooleanValue != nil {
				configVariables[param.Parameter.Key] = param.Parameter.DefaultValue.BooleanValue
			} else if param.Parameter.DefaultValue.StringArray != nil {
				configVariables[param.Parameter.Key] = param.Parameter.DefaultValue.StringArray.StringValues
			}
		}
	}
	if len(configVariables) > 0 {
		response.RespBody, response.Err = json.Marshal(configVariables)
	}
	return response
}

// Delete
func Delete(name string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name)
	return apiclient.HttpClient(u.String(), "", "DELETE")
}

// DeleteVersion
func DeleteVersion(name string, version string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)
	return apiclient.HttpClient(u.String(), "", "DELETE")
}

// DeleteByUserlabel
func DeleteByUserlabel(name string, userLabel string) apiclient.APIResponse {
	response := GetByUserlabel(name, userLabel, false, false, false)
	if response.Err != nil {
		return response
	}

	iversion := integrationVersion{}
	err := json.Unmarshal(response.RespBody, &iversion)
	if err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}

	version := getVersion(iversion.Name)
	return DeleteVersion(name, version)
}

// DeleteBySnapshot
func DeleteBySnapshot(name string, snapshot string) apiclient.APIResponse {
	response := GetBySnapshot(name, snapshot, false, false, false)
	if response.Err != nil {
		return response
	}

	iversion := integrationVersion{}
	err := json.Unmarshal(response.RespBody, &iversion)
	if err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}

	version := getVersion(iversion.Name)
	return DeleteVersion(name, version)
}

// Deactivate
func Deactivate(name string, version string) apiclient.APIResponse {
	return changeState(name, version, "", nil, ":deactivate")
}

// Archive
func Archive(name string, version string) apiclient.APIResponse {
	return changeState(name, version, "", nil, ":archive")
}

// Publish
func Publish(name string, version string, configVariables []byte) apiclient.APIResponse {
	return changeState(name, version, "", configVariables, ":publish")
}

// Unpublish
func Unpublish(name string, version string) apiclient.APIResponse {
	return changeState(name, version, "", nil, ":unpublish")
}

// UnpublishSnapshot
func UnpublishSnapshot(name string, snapshot string) apiclient.APIResponse {
	return changeState(name, "", "snapshotNumber="+snapshot, nil, ":unpublish")
}

// UnpublishUserLabel
func UnpublishUserLabel(name string, userLabel string) apiclient.APIResponse {
	return changeState(name, "", "userLabel="+userLabel, nil, ":unpublish")
}

// Download
func Download(name string, version string) apiclient.APIResponse {
	return changeState(name, version, "", nil, ":download")
}

// ArchiveSnapshot
func ArchiveSnapshot(name string, snapshot string) apiclient.APIResponse {
	return changeState(name, "", "snapshotNumber="+snapshot, nil, ":archive")
}

// DeactivateSnapshot
func DeactivateSnapshot(name string, snapshot string) apiclient.APIResponse {
	return changeState(name, "", "snapshotNumber="+snapshot, nil, ":deactivate")
}

// ArchiveUserLabel
func ArchiveUserLabel(name string, userLabel string) apiclient.APIResponse {
	return changeState(name, "", "userLabel="+userLabel, nil, ":archive")
}

// DeactivateUserLabel
func DeactivateUserLabel(name string, userLabel string) apiclient.APIResponse {
	return changeState(name, "", "userLabel="+userLabel, nil, ":deactivate")
}

// PublishUserLabel
func PublishUserLabel(name string, userlabel string, configVariables []byte) apiclient.APIResponse {
	return changeState(name, "", "userLabel="+userlabel, configVariables, ":publish")
}

// PublishSnapshot
func PublishSnapshot(name string, snapshot string, configVariables []byte) apiclient.APIResponse {
	return changeState(name, "", "snapshotNumber="+snapshot, configVariables, ":publish")
}

// DownloadSnapshot
func DownloadSnapshot(name string, snapshot string) apiclient.APIResponse {
	var version string
	var err error
	if version, err = getVersionId(name, "snapshotNumber="+snapshot); err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}
	return Download(name, version)
}

// DownloadSnapshot
func DownloadUserLabel(name string, userlabel string) apiclient.APIResponse {
	var version string
	var err error
	if version, err = getVersionId(name, "userLabel="+userlabel); err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}
	return Download(name, version)
}

// GetAuthConfigs
func GetAuthConfigs(integration []byte) (authcfgs []string, err error) {
	iversion := integrationVersion{}

	err = json.Unmarshal(integration, &iversion)
	if err != nil {
		return authcfgs, err
	}

	for _, taskConfig := range iversion.TaskConfigs {
		if taskConfig.Task == "GenericRestV2Task" || taskConfig.Task == "CloudFunctionTask" {
			authConfigParams := taskConfig.Parameters["authConfig"]
			if authConfigParams.Key == "authConfig" {
				authConfigUuid := getAuthConfigUuid(*authConfigParams.Value.JsonValue)
				if authConfigUuid != "" {
					authcfgs = append(authcfgs, authConfigUuid)
				}
			}
			authConfigNameParams := taskConfig.Parameters["authConfigName"]
			if authConfigNameParams.Key == "authConfigName" && *authConfigNameParams.Value.StringValue != "" {
				authConfigUuid, err := authconfigs.Find(*authConfigNameParams.Value.StringValue, "")
				if err != nil {
					return nil, fmt.Errorf("unable to find authconfig with name %s", *authConfigNameParams.Value.StringValue)
				}
				authcfgs = append(authcfgs, authConfigUuid)
			}
		}
	}

	return authcfgs, err
}

// GetSfdcInstances
func GetSfdcInstances(integration []byte) (instances map[string]string, err error) {
	iversion := integrationVersion{}

	err = json.Unmarshal(integration, &iversion)
	if err != nil {
		return instances, err
	}

	instances = make(map[string]string)

	for _, triggerConfig := range iversion.TriggerConfigs {
		if triggerConfig.TriggerType == "SFDC_CHANNEL" {
			instances[triggerConfig.Properties["SFDC instance name"]] = triggerConfig.Properties["Channel name"]
		}
	}

	return instances, err
}

// GetConnections
func GetConnections(integration []byte) (connections []string, err error) {
	iversion := integrationVersion{}

	err = json.Unmarshal(integration, &iversion)
	if err != nil {
		return connections, err
	}

	for _, taskConfig := range iversion.TaskConfigs {
		if taskConfig.Task == "GenericConnectorTask" {
			connectionParams := taskConfig.Parameters["config"]
			if connectionParams.Key == "config" {
				connectionName := getConnectionName(*connectionParams.Value.JsonValue)
				connections = append(connections, connectionName)
			}
		}
	}

	for _, triggerConfig := range iversion.TriggerConfigs {
		if triggerConfig.TriggerType == "INTEGRATION_CONNECTOR_TRIGGER" {
			connections = append(connections, triggerConfig.Properties["Connection name"])
		}
	}
	return connections, err
}

// GetConnectionsWithRegion
func GetConnectionsWithRegion(integration []byte) (connections []integrationConnection, err error) {
	iversion := integrationVersion{}

	err = json.Unmarshal(integration, &iversion)
	if err != nil {
		return connections, err
	}

	for _, taskConfig := range iversion.TaskConfigs {
		if taskConfig.Task == "GenericConnectorTask" {
			connectionParams := taskConfig.Parameters["config"]
			if connectionParams.Key == "config" && connectionParams.Value.JsonValue != nil {
				newConnection := integrationConnection{}
				newConnection.Name = getConnectionName(*connectionParams.Value.JsonValue)
				newConnection.Region = getConnectionRegion(*connectionParams.Value.JsonValue)
				newConnection.Version = getConnectionVersion(*connectionParams.Value.JsonValue)
				newConnection.CustomConnection = false
				connections = append(connections, newConnection)
			}
			if _, ok := taskConfig.Parameters["connectionName"]; ok {
				// check custom connection
				if isCustomConnection(taskConfig.Parameters["connectionVersion"]) {
					newCustomConnection := getIntegrationCustomConnection(taskConfig.Parameters["connectionVersion"])
					connections = append(connections, newCustomConnection)
					newConnection := getIntegrationConnection(taskConfig.Parameters["connectionName"],
						taskConfig.Parameters["connectionVersion"], iversion.IntegrationConfigParameters)
					connections = append(connections, newConnection)
				} else {
					newConnection := getIntegrationConnection(taskConfig.Parameters["connectionName"],
						taskConfig.Parameters["connectionVersion"], iversion.IntegrationConfigParameters)
					connections = append(connections, newConnection)
				}
			}
		}
	}
	for _, triggerConfig := range iversion.TriggerConfigs {
		if triggerConfig.TriggerType == "INTEGRATION_CONNECTOR_TRIGGER" {
			newConnection := integrationConnection{}
			newConnection.Name = triggerConfig.Properties["Connection name"]
			newConnection.Region = triggerConfig.Properties["Region"]
			connections = append(connections, newConnection)
		}
	}
	return connections, err
}

// GetVersion
func GetVersion(name string, userLabel string, snapshot string) (version string, err error) {
	var response apiclient.APIResponse

	if userLabel != "" {
		response = GetByUserlabel(name, userLabel, true, false, false)
		if err != nil {
			return "", err
		}
	} else if snapshot != "" {
		response = GetBySnapshot(name, snapshot, true, false, false)
		if response.Err != nil {
			return "", response.Err
		}
	} else {
		return "", fmt.Errorf("userLabel or snapshot must be passed")
	}
	return GetIntegrationVersion(response.RespBody)
}

func GetIntegrationVersion(respBody []byte) (string, error) {
	var data map[string]interface{}
	err := json.Unmarshal(respBody, &data)
	if err != nil {
		return "", err
	}
	if data["integrationVersions"] == nil {
		if data["version"] == nil {
			return "", fmt.Errorf("no integration versions were found")
		} else {
			return data["version"].(string), nil
		}
	}
	integrationVersions := data["integrationVersions"].([]interface{})
	firstIntegrationVersion := integrationVersions[0].(map[string]interface{})
	if firstIntegrationVersion["version"].(string) == "" {
		return "", fmt.Errorf("unable to extract version id from integration")
	}
	return firstIntegrationVersion["version"].(string), nil
}

// changeState
func changeState(name string, version string, filter string, configVars []byte, action string) apiclient.APIResponse {
	var err error
	var response apiclient.APIResponse

	// if a version is sent, use it, else try the filter
	if version == "" {
		if version, err = getVersionId(name, filter); err != nil {
			return apiclient.APIResponse{
				Err: err,
			}
		}
	}
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version+action)
	// download is a get, the rest are post
	if action == ":download" {
		response = apiclient.HttpClient(u.String())
	} else if action == ":publish" {
		if configVars != nil {
			contents := string(configVars)
			contents = strings.Replace(contents, "\n", "", -1)
			contents = strings.Replace(contents, "\t", "", -1)
			contents = strings.Replace(contents, "\\", "", -1)
			contents = fmt.Sprintf("{\"configParameters\":%s}", contents)
			response = apiclient.HttpClient(u.String(), contents)
		} else {
			response = apiclient.HttpClient(u.String(), "")
		}
	} else {
		response = apiclient.HttpClient(u.String(), "")
	}
	return response
}

// getVersionId
func getVersionId(name string, filter string) (version string, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	q := u.Query()
	q.Set("filter", filter)

	u.RawQuery = q.Encode()
	u.Path = path.Join(u.Path, "integrations", name, "versions")
	response := apiclient.HttpClient(u.String())
	if response.Err != nil {
		return "", err
	}

	iversions := listIntegrationVersions{}
	if err = json.Unmarshal(response.RespBody, &iversions); err != nil {
		return "", err
	}

	if len(iversions.IntegrationVersions) > 0 {
		return iversions.IntegrationVersions[0].Name[strings.LastIndex(iversions.IntegrationVersions[0].Name, "/")+1:], nil
	} else {
		return "", fmt.Errorf("filter condition not found")
	}
}

// ExportConcurrent exports all Integration Flows in the specified folder using a configurable number of connections
func ExportConcurrent(folder string, numConnections int) error {
	// Set export settings
	apiclient.SetExportToFile(folder)

	pageToken := ""
	lintegrations := listintegrations{}

	for {
		l := listintegrations{}
		response := List(maxPageSize, pageToken, "", "")
		if response.Err != nil {
			return fmt.Errorf("failed to fetch Integrations: %w", response.Err)
		}
		err := json.Unmarshal(response.RespBody, &l)
		if err != nil {
			return fmt.Errorf("failed to unmarshall: %w", err)
		}
		lintegrations.Integrations = append(lintegrations.Integrations, l.Integrations...)
		if l.NextPageToken == "" {
			break
		}
	}

	errChan := make(chan error)
	workChan := make(chan integration, len(lintegrations.Integrations))

	fanOutWg := sync.WaitGroup{}
	fanInWg := sync.WaitGroup{}

	errs := []string{}
	fanInWg.Add(1)

	go func() {
		defer fanInWg.Done()
		for {
			newErr, ok := <-errChan
			if !ok {
				return
			}
			errs = append(errs, newErr.Error())
		}
	}()

	for i := 0; i < numConnections; i++ {
		fanOutWg.Add(1)
		go exportWorker(&fanOutWg, workChan, errChan)
	}

	for _, i := range lintegrations.Integrations {
		workChan <- i
	}

	close(workChan)
	fanOutWg.Wait()
	close(errChan)
	fanInWg.Wait()

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func exportWorker(wg *sync.WaitGroup, workCh <-chan integration, errs chan<- error) {
	defer wg.Done()
	for {
		work, ok := <-workCh
		if !ok {
			return
		}
		integrationName := work.Name[strings.LastIndex(work.Name, "/")+1:]
		clilog.Info.Printf("Exporting all the revisions for Integration Flow %s\n", integrationName)

		if response := ListVersions(integrationName, maxPageSize, "", "", "", true, false, false); response.Err != nil {
			errs <- response.Err
		}
	}
}

// Export
func Export(folder string) (err error) {
	apiclient.SetExportToFile(folder)

	pageToken := ""
	lintegrations := listintegrations{}

	for {
		l := listintegrations{}
		response := List(maxPageSize, pageToken, "", "")
		if err != nil {
			return fmt.Errorf("failed to fetch Integrations: %w", response.Err)
		}
		err = json.Unmarshal(response.RespBody, &l)
		if err != nil {
			return fmt.Errorf("failed to unmarshall: %w", err)
		}
		lintegrations.Integrations = append(lintegrations.Integrations, l.Integrations...)
		pageToken = l.NextPageToken
		if l.NextPageToken == "" {
			break
		}
	}

	// no integrations where found
	if len(lintegrations.Integrations) == 0 {
		return nil
	}

	for _, lintegration := range lintegrations.Integrations {
		integrationName := lintegration.Name[strings.LastIndex(lintegration.Name, "/")+1:]
		clilog.Info.Printf("Exporting all the revisions for Integration Flow %s\n", integrationName)
		if response := ListVersions(integrationName, maxPageSize, "", "", "", true, false, false); response.Err != nil {
			return response.Err
		}
	}

	return nil
}

// ImportFlow
func ImportFlow(name string, folder string, numConnections int) (err error) {
	var versions []string

	rIntegrationFlowFiles := regexp.MustCompile(name + `\+[0-9]+\+[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}\.json`)

	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			clilog.Warning.Println("integration folder not found")
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		fileName := filepath.Base(path)
		ok := rIntegrationFlowFiles.Match([]byte(fileName))
		if ok {
			versions = append(versions, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	numEntities := len(versions)
	clilog.Info.Printf("Found %d versions for integration %s in the folder\n", numEntities, name)
	clilog.Debug.Printf("Importing versions with %d connections\n", numConnections)

	errChan := make(chan error)
	workChan := make(chan []string, numEntities)

	fanOutWg := sync.WaitGroup{}
	fanInWg := sync.WaitGroup{}

	errs := []string{}
	fanInWg.Add(1)

	go func() {
		defer fanInWg.Done()
		for {
			newErr, ok := <-errChan
			if !ok {
				return
			}
			errs = append(errs, newErr.Error())
		}
	}()

	for i := 0; i < numConnections; i++ {
		fanOutWg.Add(1)
		go batchImport(&fanOutWg, name, workChan, errChan)
	}

	workChan <- versions

	close(workChan)
	fanOutWg.Wait()
	close(errChan)
	fanInWg.Wait()

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

// importWorker
func importWorker(wg *sync.WaitGroup, workCh <-chan string, folder string, numConnections int, errs chan<- error) {
	defer wg.Done()
	for {
		work, ok := <-workCh
		if !ok {
			return
		}
		integrationFlowName := extractIntegrationFlowName(work)
		if err := uploadAsync(integrationFlowName, work); err != nil {
			errs <- err
		}
	}
}

// batchImport creates a batch of integration flows to import
func batchImport(wg *sync.WaitGroup, name string, workCh <-chan []string, errs chan<- error) {
	defer wg.Done()

	for _, work := range <-workCh {
		// could possibly extend this to use batchImport
		err := uploadAsync(name, work)
		if err != nil {
			errs <- err
			continue
		}
	}
}

func uploadAsync(name string, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if response := CreateVersion(name, content, nil, "", "", false, false); response.Err != nil {
		return response.Err
	}

	clilog.Info.Printf("Uploaded file %s for Integration flow %s\n", filePath, name)
	return nil
}

// Import
func Import(folder string, numConnections int) (err error) {
	var fileNames []string

	rIntegrationFlowFiles := regexp.MustCompile(`[\w|-]+\+[0-9]+\+[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}\.json`)

	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			clilog.Warning.Println("integration folder not found")
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		fileName := filepath.Base(path)
		if ok := rIntegrationFlowFiles.Match([]byte(fileName)); ok {
			fileNames = append(fileNames, path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	numEntities := len(fileNames)
	clilog.Info.Printf("Found %d Integration Versions in the folder\n", numEntities)
	clilog.Debug.Printf("Importing versions with %d connections\n", numConnections)

	errChan := make(chan error)
	workChan := make(chan string, numEntities)

	fanOutWg := sync.WaitGroup{}
	fanInWg := sync.WaitGroup{}

	errs := []string{}
	fanInWg.Add(1)

	go func() {
		defer fanInWg.Done()
		for {
			newErr, ok := <-errChan
			if !ok {
				return
			}
			errs = append(errs, newErr.Error())
		}
	}()

	for i := 0; i < numConnections; i++ {
		fanOutWg.Add(1)
		go importWorker(&fanOutWg, workChan, folder, numConnections, errChan)
	}

	for _, fileName := range fileNames {
		workChan <- fileName
	}

	close(workChan)
	fanOutWg.Wait()
	close(errChan)
	fanInWg.Wait()

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

// extractIntegrationFlowName
func extractIntegrationFlowName(fileName string) (name string) {
	splitNames := strings.Split(fileName, "+")
	return splitNames[0]
}

// integrationFlowExists
func integrationFlowExists(name string, integrationFlowList []string) bool {
	for _, integrationFlow := range integrationFlowList {
		if name == integrationFlow {
			return true
		}
	}
	return false
}

// asyncImportFlow
func asyncImportFlow(name string, folder string, conn int, pwg *sync.WaitGroup) {
	defer pwg.Done()

	_ = ImportFlow(name, folder, conn)
}

// getVersion
func getVersion(name string) (version string) {
	s := strings.Split(name, "/")
	return s[len(s)-1]
}

func convertInternalToExternal(internalVersion integrationVersion) (externalVersion integrationVersionExternal) {
	externalVersion = integrationVersionExternal{}
	externalVersion.Description = internalVersion.Description
	externalVersion.SnapshotNumber = internalVersion.SnapshotNumber
	externalVersion.TriggerConfigs = internalVersion.TriggerConfigs
	externalVersion.TaskConfigs = internalVersion.TaskConfigs
	externalVersion.IntegrationParameters = internalVersion.IntegrationParameters
	externalVersion.IntegrationConfigParameters = internalVersion.IntegrationConfigParameters
	if internalVersion.UserLabel != nil {
		externalVersion.UserLabel = new(string)
		*externalVersion.UserLabel = *internalVersion.UserLabel
	}
	externalVersion.ErrorCatcherConfigs = internalVersion.ErrorCatcherConfigs
	externalVersion.DatabasePersistencePolicy = internalVersion.DatabasePersistencePolicy
	externalVersion.EnableVariableMasking = internalVersion.EnableVariableMasking
	externalVersion.CloudLoggingDetails = internalVersion.CloudLoggingDetails

	return externalVersion
}

func getBasicInfo(respBody []byte) apiclient.APIResponse {
	iVer := integrationVersion{}
	bIVer := basicIntegrationVersion{}
	var err error
	var response apiclient.APIResponse

	if err = json.Unmarshal(respBody, &iVer); err != nil {
		return apiclient.APIResponse{
			Err: err,
		}
	}

	bIVer.SnapshotNumber = iVer.SnapshotNumber
	bIVer.Version = getVersion(iVer.Name)

	response.RespBody, response.Err = json.Marshal(bIVer)

	return response
}

// getAuthConfigUuid
func getAuthConfigUuid(jsonValue string) string {
	var m map[string]string
	jsonValue = strings.Replace(jsonValue, "\n", "", -1)
	_ = json.Unmarshal([]byte(jsonValue), &m)
	return m["authConfigId"]
}

// getConnectionName
func getConnectionName(jsonValue string) string {
	type connection struct {
		ConnectionName    string `json:"connectionName,omitempty"`
		ServiceName       string `json:"serviceName,omitempty"`
		ConnectionVersion string `json:"connectionVersion,omitempty"`
	}

	type config struct {
		Type       string     `json:"@type,omitempty"`
		Connection connection `json:"connection,omitempty"`
		Operation  string     `json:"operation,omitempty"`
	}

	c := config{}

	_ = json.Unmarshal([]byte(jsonValue), &c)
	name := c.Connection.ConnectionName
	return name[strings.LastIndex(name, "/")+1:]
}

// getConnectionRegion
func getConnectionRegion(jsonValue string) string {
	type connection struct {
		ConnectionName    string `json:"connectionName,omitempty"`
		ServiceName       string `json:"serviceName,omitempty"`
		ConnectionVersion string `json:"connectionVersion,omitempty"`
	}

	type config struct {
		Type       string     `json:"@type,omitempty"`
		Connection connection `json:"connection,omitempty"`
		Operation  string     `json:"operation,omitempty"`
	}

	c := config{}

	_ = json.Unmarshal([]byte(jsonValue), &c)
	name := c.Connection.ConnectionName
	r := regexp.MustCompile(`.*/locations/(.*)/connections/.*`)
	return r.FindStringSubmatch(name)[1]
}

func getConnectionVersion(jsonValue string) string {
	type connection struct {
		ConnectionName    string `json:"connectionName,omitempty"`
		ServiceName       string `json:"serviceName,omitempty"`
		ConnectionVersion string `json:"connectionVersion,omitempty"`
	}

	type config struct {
		Type       string     `json:"@type,omitempty"`
		Connection connection `json:"connection,omitempty"`
		Operation  string     `json:"operation,omitempty"`
	}

	c := config{}

	_ = json.Unmarshal([]byte(jsonValue), &c)
	version := c.Connection.ConnectionVersion
	return version[strings.LastIndex(version, "/")+1:]
}

func getJson(contents string) map[string]interface{} {
	contents = strings.Replace(contents, "\n", "", -1)
	m := make(map[string]interface{})
	json.Unmarshal([]byte(contents), &m)
	return m
}

func getIntegrationCustomConnection(connectionVersion eventparameter) integrationConnection {
	ic := integrationConnection{}
	ic.Name = strings.Split(*connectionVersion.Value.StringValue, "/")[7]
	ic.Version = strings.Split(*connectionVersion.Value.StringValue, "/")[9]
	ic.Region = "global"
	ic.CustomConnection = true
	return ic
}

func getIntegrationConnection(connectionName eventparameter,
	connectionVersion eventparameter, configParams []parameterConfig,
) integrationConnection {
	ic := integrationConnection{}

	// determine connection name.

	// connection name is a variable
	if strings.HasPrefix(*connectionName.Value.StringValue, "$`CONFIG_") {
		cName := getConfigParamValue(*connectionName.Value.StringValue, configParams)
		if cName != "" {
			ic.Name = strings.Split(cName, "/")[5]
			ic.Region = strings.Split(cName, "/")[3]
		}
	} else {
		ic.Name = strings.Split(*connectionName.Value.StringValue, "/")[5]
		ic.Region = strings.Split(*connectionName.Value.StringValue, "/")[3]
	}

	ic.Version = strings.Split(*connectionVersion.Value.StringValue, "/")[9]
	ic.CustomConnection = false
	return ic
}

func getConfigParamValue(name string, configParams []parameterConfig) string {
	name = strings.ReplaceAll(name, "$", "")
	for _, configParam := range configParams {
		if configParam.Parameter.Key == name {
			if configParam.Value != nil && configParam.Value.StringValue != nil {
				return *configParam.Value.StringValue
			} else if configParam.Parameter.DefaultValue != nil && configParam.Parameter.DefaultValue.StringValue != nil {
				return *configParam.Parameter.DefaultValue.StringValue
			}
		}
	}
	return ""
}

func isCustomConnection(connectionVersion eventparameter) bool {
	connectionType := strings.Split(*connectionVersion.Value.StringValue, "/")[5]
	if strings.EqualFold(connectionType, "customConnector") {
		return true
	} else {
		return false
	}
}

// GetInputParameters
func GetInputParameters(integrationBody []byte) (execConfig []byte, err error) {
	iversion := integrationVersionExternal{}

	inputParameters := []string{}

	const emptyTestConfig = `{
		"inputParameters": {}
}`

	err = json.Unmarshal(integrationBody, &iversion)
	if err != nil {
		return []byte(emptyTestConfig), err
	}

	for _, p := range iversion.IntegrationParameters {
		if p.InputOutputType == "IN" {
			switch p.DataType {
			case "STRING_VALUE":
				inputParameters = append(inputParameters, fmt.Sprintf("\"%s\": {\"stringValue\": \"\"}", p.Key))
			case "INT_VALUE":
				inputParameters = append(inputParameters, fmt.Sprintf("\"%s\": {\"intValue\": 0}", p.Key))
			case "BOOLEAN_VALUE":
				inputParameters = append(inputParameters, fmt.Sprintf("\"%s\": {\"booleanValue\": false}", p.Key))
			case "JSON_VALUE":
				inputParameters = append(inputParameters, fmt.Sprintf("\"%s\": {\"jsonValue\": {}}", p.Key))
			case "DOUBLE_TPYE":
				inputParameters = append(inputParameters, fmt.Sprintf("\"%s\": {\"doubleValue\": 0.0}", p.Key))
			case "INT_ARRAY":
				inputParameters = append(inputParameters, fmt.Sprintf("\"%s\": {\"intArray\": {\"intValues\": [0]}}", p.Key))
			case "STRING_ARRAY":
				inputParameters = append(inputParameters, fmt.Sprintf("\"%s\": {\"stringArray\": {\"stringValues\":[\"\"]}}", p.Key))
			}
		}
	}

	if len(inputParameters) == 0 {
		return []byte(emptyTestConfig), nil
	}

	return apiclient.PrettifyJson([]byte("{\"inputParameters\": {" + strings.Join(inputParameters, ",") + "}}"))
}
