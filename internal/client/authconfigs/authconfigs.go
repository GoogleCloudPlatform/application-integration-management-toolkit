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
	"internal/apiclient"
	"internal/clilog"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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
	CredentialType                 string                          `json:"credentialType,omitempty"`
	UsernameAndPassword            *usernameAndPassword            `json:"usernameAndPassword,omitempty"`
	OidcToken                      *oidcToken                      `json:"oidcToken,omitempty"`
	Jwt                            *jwt                            `json:"jwt,omitempty"`
	ServiceAccountCredentials      *serviceAccountCredentials      `json:"serviceAccountCredentials,omitempty"`
	AuthToken                      *authToken                      `json:"authToken,omitempty"`
	OAuth2ResourceOwnerCredentials *oauth2ResourceOwnerCredentials `json:"oauth2ResourceOwnerCredentials,omitempty"`
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

type oauth2ResourceOwnerCredentials struct {
	ClientId      string `json:"clientId,omitempty"`
	ClientSecret  string `json:"clientSecret,omitempty"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	TokenEndpoint string `json:"tokenEndpoint,omitempty"`
	RequestType   string `json:"requestType,omitempty"`
	Scope         string `json:"scope,omitempty"`
}

// Create
func Create(content []byte) apiclient.APIResponse {
	c := authConfig{}

	if err := json.Unmarshal(content, &c); err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      apiclient.NewCliError("error unmarshalling", err),
		}
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())

	u.Path = path.Join(u.Path, "authConfigs")
	return apiclient.HttpClient(u.String(), string(content))
}

// Delete
func Delete(name string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "authConfigs", name)
	return apiclient.HttpClient(u.String(), "", "DELETE")
}

// Get
func Get(name string, minimal bool) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "authConfigs", name)

	response := apiclient.HttpClient(u.String())
	if minimal {
		iversion := authConfig{}
		err := json.Unmarshal(response.RespBody, &iversion)
		if err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      apiclient.NewCliError("error unmarshalling", err),
			}
		}
		eversion := convertInternalToExternal(iversion)
		respBody, err := json.Marshal(eversion)
		if err != nil {
			return apiclient.APIResponse{
				RespBody: nil,
				Err:      apiclient.NewCliError("error marshalling", err),
			}
		}
		return apiclient.APIResponse{
			RespBody: respBody,
			Err:      nil,
		}
	}
	return response
}

// GetDisplayName
func GetDisplayName(name string) (displayName string, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "authConfigs", name)

	response := apiclient.HttpClient(u.String())
	if response.Err != nil {
		return "", err
	}

	iversion := authConfig{}
	err = json.Unmarshal(response.RespBody, &iversion)
	if err != nil {
		return "", apiclient.NewCliError("error unmarshalling", err)
	}

	return iversion.DisplayName, nil
}

// List
func List(pageSize int, pageToken string, filter string) apiclient.APIResponse {
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
	return apiclient.HttpClient(u.String())
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
	if response := apiclient.HttpClient(u.String()); response.Err != nil {
		return "", response.Err
	}

	if err = json.Unmarshal(respBody, &ac); err != nil {
		return "", apiclient.NewCliError("error unmarshalling", err)
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

	apiclient.SetExportToFile(folder)

	if response := List(100, "", ""); response.Err != nil {
		return response.Err
	}

	fileName := "authconfigs_" + strconv.Itoa(count) + ".json"
	if err = apiclient.WriteByteArrayToFile(path.Join(apiclient.GetExportToFile(), fileName), false, respBody); err != nil {
		return apiclient.NewCliError("error writing to file", err)
	}
	clilog.Info.Printf("Downloaded %s\n", fileName)

	aconfigs := authConfigs{}
	if err = json.Unmarshal(respBody, &aconfigs); err != nil {
		return err
	}

	for aconfigs.NextPageToken != "" {

		if response := List(100, "", ""); response.Err != nil {
			return response.Err
		}

		if err = json.Unmarshal(respBody, &aconfigs); err != nil {
			return apiclient.NewCliError("error unmarshalling", err)
		}

		count++
		fileName := "authconfigs_" + strconv.Itoa(count) + ".json"
		if err = apiclient.WriteByteArrayToFile(path.Join(apiclient.GetExportToFile(), fileName), false, respBody); err != nil {
			return apiclient.NewCliError("error writing to file", err)
		}
		clilog.Info.Printf("Downloaded %s\n", fileName)
	}

	return nil
}

func Patch(name string, content []byte, updateMask []string) apiclient.APIResponse {
	a := authConfig{}
	if err := json.Unmarshal(content, &a); err != nil {
		return apiclient.APIResponse{
			RespBody: nil,
			Err:      apiclient.NewCliError("error unmarshalling", err),
		}
	}

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())

	if len(updateMask) != 0 {
		updates := strings.Join(updateMask, ",")
		q := u.Query()
		q.Set("updateMask", updates)
		u.RawQuery = q.Encode()
	}

	u.Path = path.Join(u.Path, "authConfigs", name)

	return apiclient.HttpClient(u.String(), string(content), "PATCH")
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
