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
) (respBody []byte, err error) {
	iversion := integrationVersion{}
	if err = json.Unmarshal(content, &iversion); err != nil {
		return nil, err
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
			return nil, err
		}
		if eversion, err = mergeOverrides(eversion, o, grantPermission); err != nil {
			return nil, err
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
		return nil, err
	}

	if basicInfo {
		apiclient.ClientPrintHttpResponse.Set(false)
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions")
	respBody, err = apiclient.HttpClient(u.String(), string(content))

	if basicInfo {
		var respBasicBody []byte
		apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
		if respBasicBody, err = getBasicInfo(respBody); err != nil {
			return nil, err
		}
		apiclient.PrettyPrint(respBasicBody)
		return respBasicBody, nil
	}

	return respBody, err
}

// Upload
func Upload(name string, content []byte) (respBody []byte, err error) {
	uploadVersion := uploadIntegrationFormat{}
	if err = json.Unmarshal(content, &uploadVersion); err != nil {
		clilog.Error.Println("invalid format for upload. Upload must have the json field content which contains " +
			"stringified integration json and optionally the file format")
		return nil, err
	}

	if uploadVersion.Content == "" {
		return nil, fmt.Errorf("invalid format for upload. Upload must have the json field content which contains " +
			"stringified integration json and optionally the file format")
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions:upload")
	respBody, err = apiclient.HttpClient(u.String(), string(content))
	return respBody, err
}

// Patch
func Patch(name string, version string, content []byte) (respBody []byte, err error) {
	iversion := integrationVersion{}
	if err = json.Unmarshal(content, &iversion); err != nil {
		return nil, err
	}

	// remove any internal elements if exists
	eversion := convertInternalToExternal(iversion)

	if content, err = json.Marshal(eversion); err != nil {
		return nil, err
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)
	respBody, err = apiclient.HttpClient(u.String(), string(content), "PATCH")
	return respBody, err
}

// TakeOverEditLock
func TakeoverEditLock(name string, version string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)
	respBody, err = apiclient.HttpClient(u.String(), "")
	return respBody, err
}

// ListVersions
func ListVersions(name string, pageSize int, pageToken string, filter string, orderBy string,
	allVersions bool, download bool, basicInfo bool,
) (respBody []byte, err error) {
	clientPrintSetting := apiclient.ClientPrintHttpResponse.Get()

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

	if apiclient.GetExportToFile() != "" {
		apiclient.ClientPrintHttpResponse.Set(false)
		defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
	}

	if !allVersions {
		if basicInfo {
			apiclient.ClientPrintHttpResponse.Set(false)
			respBody, err = apiclient.HttpClient(u.String())
			if err != nil {
				return nil, err
			}
			apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
			listIvers := listIntegrationVersions{}
			listBIvers := listbasicIntegrationVersions{}

			if err = json.Unmarshal(respBody, &listIvers); err != nil {
				return nil, err
			}

			for _, iVer := range listIvers.IntegrationVersions {
				basicIVer := basicIntegrationVersion{}
				basicIVer.SnapshotNumber = iVer.SnapshotNumber
				basicIVer.Version = getVersion(iVer.Name)
				basicIVer.State = iVer.State
				listBIvers.BasicIntegrationVersions = append(listBIvers.BasicIntegrationVersions, basicIVer)
			}
			newResp, err := json.Marshal(listBIvers)
			if clientPrintSetting {
				apiclient.PrettyPrint(newResp)
			}
			return newResp, err
		}
		respBody, err = apiclient.HttpClient(u.String())
		if err != nil {
			return nil, err
		}
		return respBody, err
	} else {
		respBody, err = apiclient.HttpClient(u.String())
		if err != nil {
			return nil, err
		}

		iversions := listIntegrationVersions{}
		if err = json.Unmarshal(respBody, &iversions); err != nil {
			return nil, err
		}

		if apiclient.GetExportToFile() != "" {
			// Write each version to a file
			for _, iversion := range iversions.IntegrationVersions {
				var iversionBytes []byte
				if iversionBytes, err = json.Marshal(iversion); err != nil {
					return nil, err
				}
				version := iversion.Name[strings.LastIndex(iversion.Name, "/")+1:]
				fileName := strings.Join([]string{name, iversion.SnapshotNumber, version}, "+") + ".json"
				if download {
					version := iversion.Name[strings.LastIndex(iversion.Name, "/")+1:]
					payload, err := Download(name, version)
					if err != nil {
						return nil, err
					}
					if err = apiclient.WriteByteArrayToFile(
						path.Join(apiclient.GetExportToFile(), fileName),
						false,
						payload); err != nil {
						return nil, err
					}
				} else {
					if err = apiclient.WriteByteArrayToFile(
						path.Join(apiclient.GetExportToFile(), fileName),
						false,
						iversionBytes); err != nil {
						return nil, err
					}
				}
				clilog.Info.Printf("Downloaded version %s for Integration flow %s\n", version, name)
			}
		}

		// if more versions exist, repeat the process
		if iversions.NextPageToken != "" {
			if _, err = ListVersions(name, -1, iversions.NextPageToken, filter, orderBy, true, download, false); err != nil {
				return nil, err
			}
		} else {
			return nil, nil
		}
	}
	return nil, err
}

// List
func List(pageSize int, pageToken string, filter string, orderBy string) (respBody []byte, err error) {
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
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

// Get
func Get(name string, version string, basicInfo bool, minimal bool, override bool) ([]byte, error) {
	if (basicInfo && minimal) || (basicInfo && override) || (minimal && override) {
		return nil, errors.New("cannot combine basicInfo, minimal and override flags")
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)
	apiclient.ClientPrintHttpResponse.Set(false)

	respBody, err := apiclient.HttpClient(u.String())

	if !override && !minimal && !basicInfo {
		apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
		apiclient.PrettyPrint(respBody)
		return respBody, nil
	}

	iversion := integrationVersion{}
	err = json.Unmarshal(respBody, &iversion)
	if err != nil {
		return nil, err
	}

	if basicInfo {
		var respBasicBody []byte
		apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
		if respBasicBody, err = getBasicInfo(respBody); err != nil {
			return nil, err
		}
		apiclient.PrettyPrint(respBasicBody)
		return respBasicBody, nil
	}

	if minimal {
		eversion := convertInternalToExternal(iversion)
		respExtBody, err := json.Marshal(eversion)
		if err != nil {
			return nil, err
		}
		apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
		apiclient.PrettyPrint(respExtBody)
		return respExtBody, nil
	}

	if override {
		var or overrides
		var respOvrBody []byte
		if or, err = extractOverrides(iversion); err != nil {
			return nil, err
		}
		if respOvrBody, err = json.Marshal(or); err != nil {
			return nil, err
		}
		apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
		apiclient.PrettyPrint(respOvrBody)
		return respOvrBody, nil
	}
	return respBody, err
}

// GetBySnapshot
func GetBySnapshot(name string, snapshot string, basicInfo bool, minimal bool, override bool) ([]byte, error) {
	apiclient.ClientPrintHttpResponse.Set(false)

	listBody, err := ListVersions(name, -1, "", "snapshotNumber="+snapshot, "", false, false, true)
	if err != nil {
		return nil, err
	}

	apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

	listBasicVersions := listbasicIntegrationVersions{}
	err = json.Unmarshal(listBody, &listBasicVersions)
	if err != nil {
		return nil, err
	}

	if len(listBasicVersions.BasicIntegrationVersions) < 1 {
		return nil, fmt.Errorf("snapshot number was not found")
	}

	version := getVersion(listBasicVersions.BasicIntegrationVersions[0].Version)
	return Get(name, version, basicInfo, minimal, override)
}

// GetByUserlabel
func GetByUserlabel(name string, userLabel string, basicInfo bool, minimal bool, override bool) ([]byte, error) {
	apiclient.ClientPrintHttpResponse.Set(false)

	listBody, err := ListVersions(name, -1, "", "userLabel="+userLabel, "", false, false, true)
	if err != nil {
		return nil, err
	}

	apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

	listBasicVersions := listbasicIntegrationVersions{}
	err = json.Unmarshal(listBody, &listBasicVersions)
	if err != nil {
		return nil, err
	}

	if len(listBasicVersions.BasicIntegrationVersions) < 1 {
		return nil, fmt.Errorf("userLabel was not found")
	}

	version := getVersion(listBasicVersions.BasicIntegrationVersions[0].Version)
	return Get(name, version, false, minimal, override)
}

// GetConfigVariables
func GetConfigVariables(contents []byte) (respBody []byte, err error) {
	iversion := integrationVersion{}
	configVariables := make(map[string]interface{})

	err = json.Unmarshal(contents, &iversion)
	if err != nil {
		return nil, err
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
		respBody, err = json.Marshal(configVariables)
	}
	return respBody, err
}

// Delete
func Delete(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name)
	respBody, err = apiclient.HttpClient(u.String(), "", "DELETE")
	return respBody, err
}

// DeleteVersion
func DeleteVersion(name string, version string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)
	respBody, err = apiclient.HttpClient(u.String(), "", "DELETE")
	return respBody, err
}

// DeleteByUserlabel
func DeleteByUserlabel(name string, userLabel string) (respBody []byte, err error) {
	apiclient.ClientPrintHttpResponse.Set(false)
	iversionBytes, err := GetByUserlabel(name, userLabel, false, false, false)
	if err != nil {
		return nil, err
	}

	iversion := integrationVersion{}
	err = json.Unmarshal(iversionBytes, &iversion)
	if err != nil {
		return nil, err
	}

	version := getVersion(iversion.Name)
	apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
	return DeleteVersion(name, version)
}

// DeleteBySnapshot
func DeleteBySnapshot(name string, snapshot string) (respBody []byte, err error) {
	apiclient.ClientPrintHttpResponse.Set(false)
	iversionBytes, err := GetBySnapshot(name, snapshot, false, false, false)
	if err != nil {
		return nil, err
	}

	iversion := integrationVersion{}
	err = json.Unmarshal(iversionBytes, &iversion)
	if err != nil {
		return nil, err
	}

	version := getVersion(iversion.Name)
	apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
	return DeleteVersion(name, version)
}

// Deactivate
func Deactivate(name string, version string) (respBody []byte, err error) {
	return changeState(name, version, "", nil, ":deactivate")
}

// Archive
func Archive(name string, version string) (respBody []byte, err error) {
	return changeState(name, version, "", nil, ":archive")
}

// Publish
func Publish(name string, version string, configVariables []byte) (respBody []byte, err error) {
	return changeState(name, version, "", configVariables, ":publish")
}

// Unpublish
func Unpublish(name string, version string) (respBody []byte, err error) {
	return changeState(name, version, "", nil, ":unpublish")
}

// UnpublishSnapshot
func UnpublishSnapshot(name string, snapshot string) (respBody []byte, err error) {
	return changeState(name, "", "snapshotNumber="+snapshot, nil, ":unpublish")
}

// UnpublishUserLabel
func UnpublishUserLabel(name string, userLabel string) (respBody []byte, err error) {
	return changeState(name, "", "userLabel="+userLabel, nil, ":unpublish")
}

// Download
func Download(name string, version string) (respBody []byte, err error) {
	return changeState(name, version, "", nil, ":download")
}

// ArchiveSnapshot
func ArchiveSnapshot(name string, snapshot string) (respBody []byte, err error) {
	return changeState(name, "", "snapshotNumber="+snapshot, nil, ":archive")
}

// DeactivateSnapshot
func DeactivateSnapshot(name string, snapshot string) (respBody []byte, err error) {
	return changeState(name, "", "snapshotNumber="+snapshot, nil, ":deactivate")
}

// ArchiveUserLabel
func ArchiveUserLabel(name string, userLabel string) (respBody []byte, err error) {
	return changeState(name, "", "userLabel="+userLabel, nil, ":archive")
}

// DeactivateUserLabel
func DeactivateUserLabel(name string, userLabel string) (respBody []byte, err error) {
	return changeState(name, "", "userLabel="+userLabel, nil, ":deactivate")
}

// PublishUserLabel
func PublishUserLabel(name string, userlabel string, configVariables []byte) (respBody []byte, err error) {
	return changeState(name, "", "userLabel="+userlabel, configVariables, ":publish")
}

// PublishSnapshot
func PublishSnapshot(name string, snapshot string, configVariables []byte) (respBody []byte, err error) {
	return changeState(name, "", "snapshotNumber="+snapshot, configVariables, ":publish")
}

// DownloadSnapshot
func DownloadSnapshot(name string, snapshot string) (respBody []byte, err error) {
	var version string
	if version, err = getVersionId(name, "snapshotNumber="+snapshot); err != nil {
		return nil, err
	}
	return Download(name, version)
}

// DownloadSnapshot
func DownloadUserLabel(name string, userlabel string) (respBody []byte, err error) {
	var version string
	if version, err = getVersionId(name, "userLabel="+userlabel); err != nil {
		return nil, err
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
	var integrationBody []byte

	apiclient.DisableCmdPrintHttpResponse()
	defer apiclient.EnableCmdPrintHttpResponse()

	if userLabel != "" {
		integrationBody, err = GetByUserlabel(name, userLabel, true, false, false)
		if err != nil {
			return "", err
		}
	} else {
		integrationBody, err = GetBySnapshot(name, snapshot, true, false, false)
		if err != nil {
			return "", err
		}
	}
	apiclient.ClientPrintHttpResponse.Set(true)
	return GetIntegrationVersion(integrationBody)
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
func changeState(name string, version string, filter string, configVars []byte, action string) (respBody []byte, err error) {
	// if a version is sent, use it, else try the filter
	if version == "" {
		if version, err = getVersionId(name, filter); err != nil {
			return nil, err
		}
	}
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version+action)
	// download is a get, the rest are post
	if action == ":download" {
		respBody, err = apiclient.HttpClient(u.String())
	} else if action == ":publish" {
		if configVars != nil {
			contents := string(configVars)
			contents = strings.Replace(contents, "\n", "", -1)
			contents = strings.Replace(contents, "\t", "", -1)
			contents = strings.Replace(contents, "\\", "", -1)
			contents = fmt.Sprintf("{\"configParameters\":%s}", contents)
			respBody, err = apiclient.HttpClient(u.String(), contents)
		} else {
			respBody, err = apiclient.HttpClient(u.String(), "")
		}
	} else {
		respBody, err = apiclient.HttpClient(u.String(), "")
	}
	return respBody, err
}

// getVersionId
func getVersionId(name string, filter string) (version string, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	q := u.Query()
	q.Set("filter", filter)

	u.RawQuery = q.Encode()
	u.Path = path.Join(u.Path, "integrations", name, "versions")
	apiclient.ClientPrintHttpResponse.Set(false)
	respBody, err := apiclient.HttpClient(u.String())
	if err != nil {
		return "", err
	}
	apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

	iversions := listIntegrationVersions{}
	if err = json.Unmarshal(respBody, &iversions); err != nil {
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
	apiclient.ClientPrintHttpResponse.Set(false)
	defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

	pageToken := ""
	lintegrations := listintegrations{}

	for {
		l := listintegrations{}
		listRespBytes, err := List(maxPageSize, pageToken, "", "")
		if err != nil {
			return fmt.Errorf("failed to fetch Integrations: %w", err)
		}
		err = json.Unmarshal(listRespBytes, &l)
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

		if _, err := ListVersions(integrationName, maxPageSize, "", "", "", true, false, false); err != nil {
			errs <- err
		}
	}
}

// Export
func Export(folder string) (err error) {
	apiclient.SetExportToFile(folder)
	apiclient.ClientPrintHttpResponse.Set(false)
	defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

	pageToken := ""
	lintegrations := listintegrations{}

	for {
		l := listintegrations{}
		listRespBytes, err := List(maxPageSize, pageToken, "", "")
		if err != nil {
			return fmt.Errorf("failed to fetch Integrations: %w", err)
		}
		err = json.Unmarshal(listRespBytes, &l)
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
		if _, err = ListVersions(integrationName, maxPageSize, "", "", "", true, false, false); err != nil {
			return err
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

	apiclient.ClientPrintHttpResponse.Set(false)
	defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

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

	if _, err := CreateVersion(name, content, nil, "", "", false, false); err != nil {
		return err
	}

	clilog.Info.Printf("Uploaded file %s for Integration flow %s\n", filePath, name)
	return nil
}

// Import
func Import(folder string, numConnections int) (err error) {
	var fileNames []string

	rIntegrationFlowFiles := regexp.MustCompile(`[\w|-]+\+[0-9]+\+[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}\.json`)

	apiclient.ClientPrintHttpResponse.Set(false)
	defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

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

func getBasicInfo(respBody []byte) (newResp []byte, err error) {
	iVer := integrationVersion{}
	bIVer := basicIntegrationVersion{}

	if err = json.Unmarshal(respBody, &iVer); err != nil {
		return nil, err
	}

	bIVer.SnapshotNumber = iVer.SnapshotNumber
	bIVer.Version = getVersion(iVer.Name)

	if newResp, err = json.Marshal(bIVer); err != nil {
		return nil, err
	}

	return newResp, err
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
