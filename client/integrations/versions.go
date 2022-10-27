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
	"path"
	"strconv"
	"strings"

	"github.com/apigee/apigeecli/clilog"
	"github.com/srinandan/integrationcli/apiclient"
)

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
