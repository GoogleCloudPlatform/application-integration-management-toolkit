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
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/apigee/apigeecli/clilog"
	"github.com/srinandan/integrationcli/apiclient"
)

const maxPageSize = 1000

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
	TriggerConfigs                []map[string]interface{} `json:"triggerConfigs,omitempty"`
	TaskConfigs                   []taskconfig             `json:"taskConfigs,omitempty"`
	IntegrationParameters         []parameterExternal      `json:"integrationParameters,omitempty"`
	UserLabel                     string                   `json:"userLabel,omitempty"`
}

type integrationVersionExternal struct {
	Description           string                   `json:"description,omitempty"`
	SnapshotNumber        string                   `json:"snapshotNumber,omitempty"`
	TriggerConfigs        []map[string]interface{} `json:"triggerConfigs,omitempty"`
	TaskConfigs           []taskconfig             `json:"taskConfigs,omitempty"`
	IntegrationParameters []parameterExternal      `json:"integrationParameters,omitempty"`
	UserLabel             string                   `json:"userLabel,omitempty"`
}

type listbasicIntegrationVersions struct {
	BasicIntegrationVersions []basicIntegrationVersion `json:"integrationVersions,omitempty"`
	NextPageToken            string                    `json:"nextPageToken,omitempty"`
}

type basicIntegrationVersion struct {
	Version        string `json:"version,omitempty"`
	SnapshotNumber string `json:"snapshotNumber,omitempty"`
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
	Key         string `json:"key,omitempty"`
	DataType    string `json:"dataType,omitempty"`
	Name        string `json:"name,omitempty"`
	IsTransient bool   `json:"isTransient,omitempty"`
	Producer    string `json:"producer,omitempty"`
}

type producedBy struct {
	ElementType       string `json:"elementType,omitempty"`
	ElementIdentifier string `json:"elementIdentifier,omitempty"`
}

type taskconfig struct {
	Task        string                    `json:"task,omitempty"`
	TaskId      string                    `json:"taskId,omitempty"`
	Parameters  map[string]eventparameter `json:"parameters,omitempty"`
	DisplayName string                    `json:"displayName,omitempty"`
}

type eventparameter struct {
	Key   string     `json:"key,omitempty"`
	Value eventvalue `json:"value,omitempty"`
}

type eventvalue struct {
	StringValue  *string          `json:"stringValue,omitempty"`
	BooleanValue *bool            `json:"booleanValue,omitempty"`
	StringArray  *stringarraytype `json:"stringArray,omitempty"`
}

type stringarraytype struct {
	StringValues []string `json:"stringValues,omitempty"`
}

// CreateVersion
func CreateVersion(name string, content []byte, overridesContent []byte, snapshot string, userlabel string) (respBody []byte, err error) {

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
		eversion = mergeOverrides(eversion, o)
	}

	if snapshot != "" {
		eversion.SnapshotNumber = snapshot
	}

	if userlabel != "" {
		eversion.UserLabel = userlabel
	}

	if content, err = json.Marshal(eversion); err != nil {
		return nil, err
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions")
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), string(content))
	return respBody, err
}

// Upload
func Upload(name string, content string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions:upload")
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), content)
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
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), string(content), "PATCH")
	return respBody, err
}

// TakeOverEditLock
func TakeoverEditLock(name string, version string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), "")
	return respBody, err
}

// ListVersions
func ListVersions(name string, pageSize int, pageToken string, filter string, orderBy string, allVersions bool, download bool, basicInfo bool) (respBody []byte, err error) {
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
		apiclient.SetPrintOutput(false)
	}

	if !allVersions {
		if basicInfo {
			apiclient.SetPrintOutput(false)
			respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
			if err != nil {
				return nil, err
			}
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
				listBIvers.BasicIntegrationVersions = append(listBIvers.BasicIntegrationVersions, basicIVer)
			}
			newResp, err := json.Marshal(listBIvers)
			apiclient.PrettyPrint(newResp)
			return newResp, err
		}
		respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
		if err != nil {
			return nil, err
		}
		return respBody, err
	} else {
		respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
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
					if err = apiclient.WriteByteArrayToFile(path.Join(apiclient.GetExportToFile(), fileName), false, payload); err != nil {
						clilog.Error.Println(err)
						return nil, err
					}
				} else {
					if err = apiclient.WriteByteArrayToFile(path.Join(apiclient.GetExportToFile(), fileName), false, iversionBytes); err != nil {
						clilog.Error.Println(err)
						return nil, err
					}
				}
				fmt.Printf("Downloaded version %s for Integration flow %s\n", version, name)
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
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	return respBody, err
}

// Get
func Get(name string, version string, basicInfo bool) ([]byte, error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "integrations", name, "versions", version)

	if basicInfo {
		iVer := integrationVersion{}
		bIVer := basicIntegrationVersion{}

		apiclient.SetPrintOutput(false)
		respBody, err := apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(respBody, &iVer); err != nil {
			return nil, err
		}

		bIVer.SnapshotNumber = iVer.SnapshotNumber
		bIVer.Version = getVersion(iVer.Name)

		newResp, err := json.Marshal(bIVer)
		apiclient.PrettyPrint(newResp)
		return newResp, err
	}

	respBody, err := apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	return respBody, err
}

// GetBySnapshot
func GetBySnapshot(name string, snapshot string) ([]byte, error) {
	apiclient.SetPrintOutput(false)
	listBody, err := ListVersions(name, -1, "", "snapshotNumber="+snapshot, "", false, false, true)
	if err != nil {
		return nil, err
	}
	apiclient.SetPrintOutput(true)

	listBasicVersions := listbasicIntegrationVersions{}
	err = json.Unmarshal(listBody, &listBasicVersions)
	if err != nil {
		return nil, err
	}

	if len(listBasicVersions.BasicIntegrationVersions) < 1 {
		return nil, fmt.Errorf("snapshot number was not found")
	}

	version := getVersion(listBasicVersions.BasicIntegrationVersions[0].Version)
	return Get(name, version, false)
}

// GetByUserlabel
func GetByUserlabel(name string, userLabel string) ([]byte, error) {
	apiclient.SetPrintOutput(false)
	listBody, err := ListVersions(name, -1, "", "userLabel="+userLabel, "", false, false, true)
	if err != nil {
		return nil, err
	}
	apiclient.SetPrintOutput(true)
	listBasicVersions := listbasicIntegrationVersions{}
	err = json.Unmarshal(listBody, &listBasicVersions)
	if err != nil {
		return nil, err
	}

	if len(listBasicVersions.BasicIntegrationVersions) < 1 {
		return nil, fmt.Errorf("userLabel was not found")
	}

	version := getVersion(listBasicVersions.BasicIntegrationVersions[0].Version)
	return Get(name, version, false)
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
func ArchiveUserLabel(name string, snapshot string) (respBody []byte, err error) {
	return changeState(name, "", "userLabel="+snapshot, ":archive")
}

// DeactivateUserLabel
func DeactivateUserLabel(name string, snapshot string) (respBody []byte, err error) {
	return changeState(name, "", "userLabel="+snapshot, ":deactivate")
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
		respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	} else {
		respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), "")
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
	apiclient.SetPrintOutput(false)
	respBody, err := apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	if err != nil {
		clilog.Error.Println(err)
		return "", err
	}
	apiclient.SetPrintOutput(true)

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

// Export
func Export(folder string) (err error) {

	apiclient.SetExportToFile(folder)
	apiclient.SetPrintOutput(false)

	respBody, err := List(maxPageSize, "", "", "")

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
		fmt.Printf("Exporting all the revisions for Integration Flow %s\n", integrationName)
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
	clilog.Info.Printf("Found %d versions in the folder\n", numEntities)
	clilog.Info.Printf("Importing versions with %d connections\n", conn)

	numOfLoops, remaining := numEntities/conn, numEntities%conn

	//ensure connections aren't greater than entities
	if conn > numEntities {
		conn = numEntities
	}

	start := 0

	apiclient.SetPrintOutput(false)

	for i, end := 0, 0; i < numOfLoops; i++ {
		pwg.Add(1)
		end = (i * conn) + conn
		clilog.Info.Printf("Uploading batch %d of versions\n", (i + 1))
		go batchImport(name, entities[start:end], &pwg)
		start = end
		pwg.Wait()
	}

	if remaining > 0 {
		pwg.Add(1)
		clilog.Info.Printf("Uploading remaining %d versions\n", remaining)
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
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	if _, err := Upload(name, string(content)); err != nil {
		clilog.Error.Println(err)
	} else {
		fmt.Printf("Uploaded file %s for Integration flow %s\n", filePath, name)
	}
}

// Import
func Import(folder string, conn int) (err error) {

	var pwg sync.WaitGroup
	var names []string

	rIntegrationFlowFiles := regexp.MustCompile(`[\w|-]+\+[0-9]+\+[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}\.json`)

	apiclient.SetPrintOutput(false)

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
			fmt.Println(integrationFlowName)
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
		//_ = ImportFlow(integrationFlowName, folder, conn)
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
	externalVersion.UserLabel = internalVersion.UserLabel
	return externalVersion
}
