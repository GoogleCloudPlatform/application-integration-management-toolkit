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

package authconfigs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"

	"github.com/apigee/apigeecli/clilog"
	"github.com/srinandan/integrationcli/apiclient"
)

type authConfigs struct {
	AuthConfig    []authConfig `json:"authConfigs,omitempty"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

type authConfig struct {
	Name                string `json:"name,omitempty"`
	DisplayName         string `json:"displayName,omitempty"`
	Description         string `json:"description,omitempty"`
	EncryptedCredential string `json:"encryptedCredential,omitempty"`
	CredentialType      string `json:"credentialType,omitempty"`
	CreatorEmail        string `json:"creatorEmail,omitempty"`
	CreateTime          string `json:"createTime,omitempty"`
	LastModifierEmail   string `json:"lastModifierEmail,omitempty"`
	Visibility          string `json:"visibility,omitempty"`
	State               string `json:"state,omitempty"`
	Reason              string `json:"reason,omitempty"`
	ValidTime           string `json:"validTime,omitempty"`
}

//Create
func Create(name string, content []byte, clientCertificate string) (respBody []byte, err error) {

	aconfig := authConfig{}
	if err = json.Unmarshal(content, aconfig); err != nil {
		return nil, err
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	q := u.Query()
	if clientCertificate != "" {
		q.Set("clientCertificate", clientCertificate)
		u.RawQuery = q.Encode()
	}

	u.Path = path.Join(u.Path, "authConfigs")
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), string(content))
	return respBody, err
}

//Delete
func Delete(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "authConfigs", name)
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), "", "DELETE")
	return respBody, err
}

//Get
func Get(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "authConfigs", name)
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	return respBody, err
}

//List
func List(pageSize int, pageToken string, filter string) (respBody []byte, err error) {
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

	u.RawQuery = q.Encode()
	u.Path = path.Join(u.Path, "authConfigs")
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	return respBody, err
}

//Export
func Export(folder string) (err error) {

	var respBody []byte
	count := 1

	apiclient.SetPrintOutput(false)
	apiclient.SetExportToFile(folder)

	if respBody, err = List(100, "", ""); err != nil {
		return err
	}

	fileName := "authconfigs_" + strconv.Itoa(count) + ".json"
	if err = apiclient.WriteByteArrayToFile(path.Join(apiclient.GetExportToFile(), fileName), false, respBody); err != nil {
		clilog.Error.Println(err)
		return err
	}
	fmt.Printf("Downloaded %s\n", fileName)

	aconfigs := authConfigs{}
	if err = json.Unmarshal(respBody, &aconfigs); err != nil {
		return err
	}

	for aconfigs.NextPageToken != "" {

		if respBody, err = List(100, "", ""); err != nil {
			return err
		}

		if err = json.Unmarshal(respBody, &aconfigs); err != nil {
			return err
		}

		count++
		fileName := "authconfigs_" + strconv.Itoa(count) + ".json"
		if err = apiclient.WriteByteArrayToFile(path.Join(apiclient.GetExportToFile(), fileName), false, respBody); err != nil {
			clilog.Error.Println(err)
			return err
		}
		fmt.Printf("Downloaded %s\n", fileName)
	}

	return nil
}
