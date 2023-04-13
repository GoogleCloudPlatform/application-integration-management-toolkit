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
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"internal/apiclient"

	"internal/clilog"
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
	UserLabel                     *string                  `json:"userLabel,omitempty"`
}

type integrationVersionExternal struct {
	Description           string              `json:"description,omitempty"`
	SnapshotNumber        string              `json:"snapshotNumber,omitempty"`
	TriggerConfigs        []triggerconfig     `json:"triggerConfigs,omitempty"`
	TaskConfigs           []taskconfig        `json:"taskConfigs,omitempty"`
	IntegrationParameters []parameterExternal `json:"integrationParameters,omitempty"`
	UserLabel             *string             `json:"userLabel,omitempty"`
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
	SynchronousCallFailurePolicy *failurePolicy            `json:"synchronousCallFailurePolicy,omitempty"`
}

type eventparameter struct {
	Key   string    `json:"key,omitempty"`
	Value valueType `json:"value,omitempty"`
}

type valueType struct {
	StringValue  *string          `json:"stringValue,omitempty"`
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
}

type cloudSchedulerConfig struct {
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty"`
	CronTab             string `json:"cronTab,omitempty"`
	Location            string `json:"location,omitempty"`
	ErrorMessage        string `json:"errorMessage,omitempty"`
}

// CreateVersion
func CreateVersion(name string, content []byte, overridesContent []byte, snapshot string,
	userlabel string) (respBody []byte, err error) {

	iversion := integrationVersion{}
	if err = json.Unmarshal(content, &iversion); err != nil {
		return nil, err
	}

	//remove any internal elements if exists
	eversion := convertInternalToExternal(iversion)

	//merge overrides if overrides were provided
	if len(overridesContent) > 0 {
		o := overrides{}
		if err = json.Unmarshal(overridesContent, &o); err != nil {
			return nil, err
		}
		if eversion, err = mergeOverrides(eversion, o); err != nil {
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

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions")

	respBody, err = apiclient.HttpClient(u.String(), string(content))
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

	//remove any internal elements if exists
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
	allVersions bool, download bool, basicInfo bool) (respBody []byte, err error) {

	clientPrintSetting := apiclient.GetClientPrintHttpResponseSetting()

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
		apiclient.SetClientPrintHttpResponse(false)
		defer apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())
	}

	if !allVersions {
		if basicInfo {
			apiclient.SetClientPrintHttpResponse(false)
			respBody, err = apiclient.HttpClient(u.String())
			if err != nil {
				return nil, err
			}
			apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())
			listIvers := listIntegrationVersions{}
			listBIvers := listbasicIntegrationVersions{}

			listBIvers.NextPageToken = listIvers.NextPageToken

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
			clilog.Error.Println(err)
			return nil, err
		}

		if apiclient.GetExportToFile() != "" {
			//Write each version to a file
			for _, iversion := range iversions.IntegrationVersions {
				var iversionBytes []byte
				if iversionBytes, err = json.Marshal(iversion); err != nil {
					clilog.Error.Println(err)
					return nil, err
				}
				version := iversion.Name[strings.LastIndex(iversion.Name, "/")+1:]
				fileName := strings.Join([]string{name, iversion.SnapshotNumber, version}, "+") + ".json"
				if download {
					version := iversion.Name[strings.LastIndex(iversion.Name, "/")+1:]
					payload, err := Download(name, version)
					if err != nil {
						clilog.Error.Println(err)
						return nil, err
					}
					if err = apiclient.WriteByteArrayToFile(
						path.Join(apiclient.GetExportToFile(), fileName),
						false,
						payload); err != nil {
						clilog.Error.Println(err)
						return nil, err
					}
				} else {
					if err = apiclient.WriteByteArrayToFile(
						path.Join(apiclient.GetExportToFile(), fileName),
						false,
						iversionBytes); err != nil {
						clilog.Error.Println(err)
						return nil, err
					}
				}
				clilog.Info.Printf("Downloaded version %s for Integration flow %s\n", version, name)
			}
		}

		//if more versions exist, repeat the process
		if iversions.NextPageToken != "" {
			if _, err = ListVersions(name, -1, iversions.NextPageToken, filter, orderBy, true, download, false); err != nil {
				clilog.Error.Println(err)
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
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)
	if basicInfo {
		apiclient.SetClientPrintHttpResponse(false)
		respBody, err := apiclient.HttpClient(u.String())

		if err != nil {
			return nil, err
		}
		//restore print setting
		apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())
		return getBasicInfo(respBody)
	}

	if override || minimal {
		apiclient.SetClientPrintHttpResponse(false)
	}

	respBody, err := apiclient.HttpClient(u.String())

	if minimal {
		iversion := integrationVersion{}
		err = json.Unmarshal(respBody, &iversion)
		if err != nil {
			return nil, err
		}

		eversion := convertInternalToExternal(iversion)
		respBody, err = json.Marshal(eversion)
		apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())
		apiclient.PrettyPrint(respBody)
	}

	if override {
		iversion := integrationVersion{}
		err = json.Unmarshal(respBody, &iversion)
		if err != nil {
			return nil, err
		}

		var or overrides

		if or, err = extractOverrides(iversion); err != nil {
			return nil, err
		}
		if respBody, err = json.Marshal(or); err != nil {
			return nil, err
		}
		apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())
		apiclient.PrettyPrint(respBody)
	}
	return respBody, err
}

// GetBySnapshot
func GetBySnapshot(name string, snapshot string, minimal bool, override bool) ([]byte, error) {

	apiclient.SetClientPrintHttpResponse(false)

	listBody, err := ListVersions(name, -1, "", "snapshotNumber="+snapshot, "", false, false, true)
	if err != nil {
		return nil, err
	}

	listBasicVersions := listbasicIntegrationVersions{}
	err = json.Unmarshal(listBody, &listBasicVersions)
	if err != nil {
		return nil, err
	}

	if len(listBasicVersions.BasicIntegrationVersions) < 1 {
		return nil, fmt.Errorf("snapshot number was not found")
	}

	version := getVersion(listBasicVersions.BasicIntegrationVersions[0].Version)
	respBody, err := Get(name, version, false, minimal, override)
	apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())
	return respBody, err
}

// GetByUserlabel
func GetByUserlabel(name string, userLabel string, minimal bool, override bool) ([]byte, error) {

	apiclient.SetClientPrintHttpResponse(false)

	listBody, err := ListVersions(name, -1, "", "userLabel="+userLabel, "", false, false, true)
	if err != nil {
		return nil, err
	}

	apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())

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

	apiclient.SetClientPrintHttpResponse(false)
	iversionBytes, err := GetByUserlabel(name, userLabel, false, false)
	if err != nil {
		return nil, err
	}

	iversion := integrationVersion{}
	err = json.Unmarshal(iversionBytes, &iversion)
	if err != nil {
		return nil, err
	}

	version := getVersion(iversion.Name)
	apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())
	return DeleteVersion(name, version)
}

// DeleteBySnapshot
func DeleteBySnapshot(name string, snapshot string) (respBody []byte, err error) {

	apiclient.SetClientPrintHttpResponse(false)
	iversionBytes, err := GetBySnapshot(name, snapshot, false, false)
	if err != nil {
		return nil, err
	}

	iversion := integrationVersion{}
	err = json.Unmarshal(iversionBytes, &iversion)
	if err != nil {
		return nil, err
	}

	version := getVersion(iversion.Name)
	apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())
	return DeleteVersion(name, version)
}

// Deactivate
func Deactivate(name string, version string) (respBody []byte, err error) {
	return changeState(name, version, "", ":deactivate")
}

// Archive
func Archive(name string, version string) (respBody []byte, err error) {
	return changeState(name, version, "", ":archive")
}

// Publish
func Publish(name string, version string) (respBody []byte, err error) {
	return changeState(name, version, "", ":publish")
}

// Unpublish
func Unpublish(name string, version string) (respBody []byte, err error) {
	return changeState(name, version, "", ":unpublish")
}

// UnpublishSnapshot
func UnpublishSnapshot(name string, snapshot string) (respBody []byte, err error) {
	return changeState(name, "", "snapshotNumber="+snapshot, ":unpublish")
}

// UnpublishUserLabel
func UnpublishUserLabel(name string, userLabel string) (respBody []byte, err error) {
	return changeState(name, "", "userLabel="+userLabel, ":unpublish")
}

// Download
func Download(name string, version string) (respBody []byte, err error) {
	return changeState(name, version, "", ":download")
}

// ArchiveSnapshot
func ArchiveSnapshot(name string, snapshot string) (respBody []byte, err error) {
	return changeState(name, "", "snapshotNumber="+snapshot, ":archive")
}

// DeactivateSnapshot
func DeactivateSnapshot(name string, snapshot string) (respBody []byte, err error) {
	return changeState(name, "", "snapshotNumber="+snapshot, ":deactivate")
}

// ArchiveUserLabel
func ArchiveUserLabel(name string, userLabel string) (respBody []byte, err error) {
	return changeState(name, "", "userLabel="+userLabel, ":archive")
}

// DeactivateUserLabel
func DeactivateUserLabel(name string, userLabel string) (respBody []byte, err error) {
	return changeState(name, "", "userLabel="+userLabel, ":deactivate")
}

// PublishUserLabel
func PublishUserLabel(name string, userlabel string) (respBody []byte, err error) {
	return changeState(name, "", "userLabel="+userlabel, ":publish")
}

// PublishSnapshot
func PublishSnapshot(name string, snapshot string) (respBody []byte, err error) {
	return changeState(name, "", "snapshotNumber="+snapshot, ":publish")
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
func GetAuthConfigs(integration []byte) (authconfigs []string, err error) {
	iversion := integrationVersion{}

	err = json.Unmarshal(integration, &iversion)
	if err != nil {
		return authconfigs, err
	}

	for _, taskConfig := range iversion.TaskConfigs {
		if taskConfig.Task == "GenericRestV2Task" {
			authConfigParams := taskConfig.Parameters["authConfig"]
			if authConfigParams.Key == "authConfig" {
				authConfigUuid := getAuthConfigUuid(*authConfigParams.Value.JsonValue)
				authconfigs = append(authconfigs, authConfigUuid)
			}
		} else if taskConfig.Task == "CloudFunctionTask" {
			authConfigParams := taskConfig.Parameters["authConfig"]
			if authConfigParams.Key == "authConfig" {
				authConfigUuid := getAuthConfigUuid(*authConfigParams.Value.JsonValue)
				authconfigs = append(authconfigs, authConfigUuid)
			}
		}
	}

	return authconfigs, err
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
	return connections, err
}

// changeState
func changeState(name string, version string, filter string, action string) (respBody []byte, err error) {
	//if a version is sent, use it, else try the filter
	if version == "" {
		if version, err = getVersionId(name, filter); err != nil {
			return nil, err
		}
	}
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version+action)
	//download is a get, the rest are post
	if action == ":download" {
		respBody, err = apiclient.HttpClient(u.String())
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
	apiclient.SetClientPrintHttpResponse(false)
	respBody, err := apiclient.HttpClient(u.String())
	if err != nil {
		clilog.Error.Println(err)
		return "", err
	}
	apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())

	iversions := listIntegrationVersions{}
	if err = json.Unmarshal(respBody, &iversions); err != nil {
		clilog.Error.Println(err)
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
	apiclient.SetClientPrintHttpResponse(false)
	defer apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())

	var pwg sync.WaitGroup
	// Build integration URL with max page size
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	q := u.Query()
	q.Set("pageSize", strconv.Itoa(maxPageSize))
	u.RawQuery = q.Encode()
	u.Path = path.Join(u.Path, "integrations")

	// Fetch first page of integrations
	lintegrations, err := fetchIntegrations(u.String())
	if err != nil {
		return err
	}

	// Create channels for work and results
	workCh := make(chan *integrationInfo, len(lintegrations.Integrations))
	resultCh := make(chan error, len(lintegrations.Integrations))

	// Start worker goroutines
	for i := 0; i < numConnections; i++ {
		pwg.Add(1)

		go exportWorker(workCh, resultCh, &pwg)
	}

	// Add integrations to work channel
	for _, lintegration := range lintegrations.Integrations {
		workCh <- &integrationInfo{
			Name: lintegration.Name,
			Path: folder,
		}
	}

	// Fetch remaining pages of integrations and add to work channel
	for lintegrations.NextPageToken != "" {
		lintegrations, err = fetchIntegrationsWithPageToken(u.String(), lintegrations.NextPageToken)
		if err != nil {
			return err
		}
		for _, lintegration := range lintegrations.Integrations {
			workCh <- &integrationInfo{
				Name: lintegration.Name,
				Path: folder,
			}
		}
	}

	// Close the work channel to signal workers to exit when all work is done
	close(workCh)

	// Collect results from result channel
	for i := 0; i < len(lintegrations.Integrations); i++ {
		if err := <-resultCh; err != nil {
			return err
		}
	}

	close(resultCh)

	pwg.Wait()

	return nil
}

func exportWorker(workCh <-chan *integrationInfo, resultCh chan<- error, pwg *sync.WaitGroup) {
	defer pwg.Done()
	for work := range workCh {
		integrationName := work.Name[strings.LastIndex(work.Name, "/")+1:]
		clilog.Info.Printf("Exporting all the revisions for Integration Flow %s\n", integrationName)

		if _, err := ListVersions(integrationName, -1, "", "", "", true, true, false); err != nil {
			resultCh <- err
		}
		resultCh <- nil
	}

}

// fetchIntegrations fetches the first page of integrations from the integration API
func fetchIntegrations(integrationURL string) (*listintegrations, error) {
	respBody, err := apiclient.HttpClient(integrationURL)
	if err != nil {
		return nil, err
	}

	var lintegrations listintegrations
	if err := json.Unmarshal(respBody, &lintegrations); err != nil {
		return nil, err
	}

	return &lintegrations, nil
}

// fetchIntegrationsWithPageToken fetches a page of integrations from the integration API using a page token
func fetchIntegrationsWithPageToken(integrationURL string, pageToken string) (*listintegrations, error) {
	u, _ := url.Parse(integrationURL)
	q := u.Query()
	q.Set("pageSize", strconv.Itoa(maxPageSize))
	q.Set("pageToken", pageToken)
	u.RawQuery = q.Encode()

	respBody, err := apiclient.HttpClient(u.String())
	if err != nil {
		return nil, err
	}

	var lintegrations listintegrations
	if err := json.Unmarshal(respBody, &lintegrations); err != nil {
		return nil, err
	}

	return &lintegrations, nil
}

// exportWorker exports all revisions

// Export
func Export(folder string) (err error) {

	apiclient.SetExportToFile(folder)
	apiclient.SetClientPrintHttpResponse(false)
	defer apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())

	respBody, err := List(maxPageSize, "", "", "")
	if err != nil {
		return err
	}

	lintegrations := listintegrations{}

	if err = json.Unmarshal(respBody, &lintegrations); err != nil {
		return err
	}

	//no integrations where found
	if len(lintegrations.Integrations) == 0 {
		return nil
	}

	for _, lintegration := range lintegrations.Integrations {
		integrationName := lintegration.Name[strings.LastIndex(lintegration.Name, "/")+1:]
		clilog.Info.Printf("Exporting all the revisions for Integration Flow %s\n", integrationName)
		if _, err = ListVersions(integrationName, -1, "", "", "", true, true, false); err != nil {
			return err
		}
	}

	if lintegrations.NextPageToken != "" {
		if err = batchExport(folder, lintegrations.NextPageToken); err != nil {
			return err
		}
	}
	return nil
}

// batchExport
func batchExport(folder string, nextPageToken string) (err error) {
	respBody, err := List(maxPageSize, nextPageToken, "", "")
	if err != nil {
		return err
	}

	lintegrations := listintegrations{}
	if err = json.Unmarshal(respBody, &lintegrations); err != nil {
		return err
	}

	//no integrations where found
	if len(lintegrations.Integrations) == 0 {
		return nil
	}

	for _, lintegration := range lintegrations.Integrations {
		integrationName := lintegration.Name[strings.LastIndex(lintegration.Name, "/")+1:]
		clilog.Debug.Printf("Exporting all the revisions for Integration Flow %s\n", integrationName)
		if _, err = ListVersions(integrationName, -1, "", "", "", true, true, false); err != nil {
			return err
		}
	}

	if lintegrations.NextPageToken != "" {
		if err = batchExport(folder, lintegrations.NextPageToken); err != nil {
			return err
		}
	}
	return nil
}

// ImportFlow
func ImportFlow(name string, folder string, conn int) (err error) {

	var pwg sync.WaitGroup
	var entities []string

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
			entities = append(entities, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	numEntities := len(entities)
	clilog.Debug.Printf("Found %d versions in the folder\n", numEntities)
	clilog.Debug.Printf("Importing versions with %d connections\n", conn)

	numOfLoops, remaining := numEntities/conn, numEntities%conn

	//ensure connections aren't greater than entities
	if conn > numEntities {
		conn = numEntities
	}

	start := 0

	apiclient.SetClientPrintHttpResponse(false)
	defer apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())

	for i, end := 0, 0; i < numOfLoops; i++ {
		pwg.Add(1)
		end = (i * conn) + conn
		clilog.Debug.Printf("Uploading batch %d of versions\n", (i + 1))
		go batchImport(name, entities[start:end], &pwg)
		start = end
		pwg.Wait()
	}

	if remaining > 0 {
		pwg.Add(1)
		clilog.Debug.Printf("Uploading remaining %d versions\n", remaining)
		go batchImport(name, entities[start:numEntities], &pwg)
		pwg.Wait()
	}

	return nil
}

// batchImport creates a batch of integration flows to import
func batchImport(name string, entities []string, pwg *sync.WaitGroup) {

	defer pwg.Done()
	//batch workgroup
	var bwg sync.WaitGroup

	bwg.Add(len(entities))

	for _, entity := range entities {
		go uploadAsync(name, entity, &bwg)
	}
	bwg.Wait()
}

func uploadAsync(name string, filePath string, wg *sync.WaitGroup) {
	defer wg.Done()
	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	if _, err := Upload(name, content); err != nil {
		clilog.Error.Println(err)
	} else {
		clilog.Info.Printf("Uploaded file %s for Integration flow %s\n", filePath, name)
	}
}

// Import
func Import(folder string, conn int) (err error) {

	var pwg sync.WaitGroup
	var names []string

	rIntegrationFlowFiles := regexp.MustCompile(`[\w|-]+\+[0-9]+\+[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}\.json`)

	apiclient.SetClientPrintHttpResponse(false)
	defer apiclient.SetClientPrintHttpResponse(apiclient.GetCmdPrintHttpResponseSetting())

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

		//collect all the flow names once
		if ok {
			integrationFlowName := extractIntegrationFlowName(fileName)
			if !integrationFlowExists(integrationFlowName, names) {
				names = append(names, integrationFlowName)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	for _, integrationFlowName := range names {
		pwg.Add(1)
		go asyncImportFlow(integrationFlowName, folder, conn, &pwg)
		pwg.Wait()
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
	if internalVersion.UserLabel != nil {
		externalVersion.UserLabel = new(string)
		*externalVersion.UserLabel = *internalVersion.UserLabel
	}
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
		apiclient.PrettyPrint(newResp)
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
