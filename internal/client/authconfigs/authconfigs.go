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
	"path/filepath"
	"strconv"

	"internal/apiclient"

	"internal/clilog"
)

type authConfigs struct {
	AuthConfig    []authConfig `json:"authConfigs,omitempty"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

type authConfig struct {
	Name                string               `json:"name,omitempty"`
	DisplayName         string               `json:"displayName,omitempty"`
	Description         string               `json:"description,omitempty"`
	EncryptedCredential *string              `json:"encryptedCredential,omitempty"`
	DecryptedCredential *decryptedCredential `json:"decryptedCredential,omitempty"`
	CreatorEmail        string               `json:"creatorEmail,omitempty"`
	CreateTime          string               `json:"createTime,omitempty"`
	LastModifierEmail   string               `json:"lastModifierEmail,omitempty"`
	Visibility          string               `json:"visibility,omitempty"`
	State               string               `json:"state,omitempty"`
	Reason              string               `json:"reason,omitempty"`
	ValidTime           string               `json:"validTime,omitempty"`
}

type authConfigExternal struct {
	DisplayName         string               `json:"displayName,omitempty"`
	Description         string               `json:"description,omitempty"`
	Visibility          string               `json:"visibility,omitempty"`
	DecryptedCredential *decryptedCredential `json:"decryptedCredential,omitempty"`
}

type decryptedCredential struct {
	CredentialType            string                     `json:"credentialType,omitempty"`
	UsernameAndPassword       *usernameAndPassword       `json:"usernameAndPassword,omitempty"`
	OidcToken                 *oidcToken                 `json:"oidcToken,omitempty"`
	Jwt                       *jwt                       `json:"jwt,omitempty"`
	ServiceAccountCredentials *serviceAccountCredentials `json:"serviceAccountCredentials,omitempty"`
	AuthToken                 *authToken                 `json:"authToken,omitempty"`
}

type usernameAndPassword struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type oidcToken struct {
	ServiceAccountEmail string `json:"serviceAccountEmail,omitempty"`
	Audience            string `json:"audience,omitempty"`
}

type jwt struct {
	JwtHeader  string `json:"jwtHeader,omitempty"`
	JwtPayload string `json:"jwtPayload,omitempty"`
	Secret     string `json:"secret,omitempty"`
}

type serviceAccountCredentials struct {
	ServiceAccount string `json:"serviceAccount,omitempty"`
	Scope          string `json:"scope,omitempty"`
}

type authToken struct {
	Type  string `json:"type,omitempty"`
	Token string `json:"token,omitempty"`
}

// Create
func Create(content []byte) (respBody []byte, err error) {
	c := authConfig{}

	if err = json.Unmarshal(content, &c); err != nil {
		return nil, err
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())

	u.Path = path.Join(u.Path, "authConfigs")
	respBody, err = apiclient.HttpClient(u.String(), string(content))
	return respBody, err
}

// Delete
func Delete(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "authConfigs", name)
	respBody, err = apiclient.HttpClient(u.String(), "", "DELETE")
	return respBody, err
}

// Get
func Get(name string, minimal bool) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "authConfigs", name)

	if minimal {
		apiclient.ClientPrintHttpResponse.Set(false)
	}
	respBody, err = apiclient.HttpClient(u.String())
	if minimal {
		iversion := authConfig{}
		err := json.Unmarshal(respBody, &iversion)
		if err != nil {
			return nil, err
		}
		eversion := convertInternalToExternal(iversion)
		respBody, err = json.Marshal(eversion)
		if err != nil {
			return nil, err
		}
		apiclient.PrettyPrint(respBody)
	}
	apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
	return respBody, err
}

// GetDisplayName
func GetDisplayName(name string) (displayName string, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "authConfigs", name)

	apiclient.ClientPrintHttpResponse.Set(false)
	defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

	respBody, err := apiclient.HttpClient(u.String())
	if err != nil {
		return "", err
	}

	iversion := authConfig{}
	err = json.Unmarshal(respBody, &iversion)
	if err != nil {
		return "", err
	}

	return iversion.DisplayName, nil
}

// List
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
	respBody, err = apiclient.HttpClient(u.String())
	return respBody, err
}

// Find
func Find(name string, pageToken string) (version string, err error) {
	ac := authConfigs{}
	var respBody []byte

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	if pageToken != "" {
		q := u.Query()
		q.Set("pageToken", pageToken)
		u.RawQuery = q.Encode()
	}

	u.Path = path.Join(u.Path, "authConfigs")
	if respBody, err = apiclient.HttpClient(u.String()); err != nil {
		return "", err
	}

	if err = json.Unmarshal(respBody, &ac); err != nil {
		return "", err
	}

	for _, config := range ac.AuthConfig {
		if config.DisplayName == name {
			version = filepath.Base(config.Name)
			return version, nil
		}
	}
	if ac.NextPageToken != "" {
		return Find(name, ac.NextPageToken)
	}
	return "", fmt.Errorf("authConfig not found")
}

// Export
func Export(folder string) (err error) {
	var respBody []byte
	count := 1

	apiclient.ClientPrintHttpResponse.Set(false)
	defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

	apiclient.SetExportToFile(folder)

	if respBody, err = List(100, "", ""); err != nil {
		return err
	}

	fileName := "authconfigs_" + strconv.Itoa(count) + ".json"
	if err = apiclient.WriteByteArrayToFile(path.Join(apiclient.GetExportToFile(), fileName), false, respBody); err != nil {
		clilog.Error.Println(err)
		return err
	}
	clilog.Info.Printf("Downloaded %s\n", fileName)

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
		clilog.Info.Printf("Downloaded %s\n", fileName)
	}

	return nil
}

// convertInternalToExternal
func convertInternalToExternal(internalVersion authConfig) (externalVersion authConfigExternal) {
	externalVersion = authConfigExternal{}
	externalVersion.DisplayName = internalVersion.DisplayName
	externalVersion.Description = internalVersion.Description
	externalVersion.Visibility = internalVersion.Visibility
	externalVersion.DecryptedCredential = new(decryptedCredential)
	externalVersion.DecryptedCredential = internalVersion.DecryptedCredential
	return externalVersion
}
