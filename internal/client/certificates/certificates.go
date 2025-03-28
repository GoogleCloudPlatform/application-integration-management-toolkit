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

package certificates

import (
	"encoding/json"
	"fmt"
	"internal/apiclient"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type certs struct {
	Cert []cert `json:"certificates,omitempty"`
}

type cert struct {
	Name              string `json:"name,omitempty"`
	DisplayName       string `json:"displayName,omitempty"`
	Description       string `json:"description,omitempty"`
	RequestorId       string `json:"requestorId,omitempty"`
	CredentialId      string `json:"credentialId,omitempty"`
	CertificateStatus string `json:"certificateStatus,omitempty"`
	ValidStartTime    string `json:"validStartTime,omitempty"`
	ValidEndTime      string `json:"validEndTime,omitempty"`
}

// Create
func Create(displayName string, description string, sslCertificate string, privateKey string, passphrase string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	certStr := []string{}
	rawCertStr := []string{}

	certStr = append(certStr, "\"displayName\":\""+displayName+"\"")
	if description != "" {
		certStr = append(certStr, "\"description\":\""+description+"\"")
	}

	rawCertStr = append(rawCertStr, "\"sslCertificate\":\""+getStringyfiedContents(sslCertificate)+"\"")
	if privateKey != "" {
		rawCertStr = append(rawCertStr, "\"encryptedPrivateKey\":\""+getStringyfiedContents(privateKey)+"\"")
	}
	if passphrase != "" {
		rawCertStr = append(rawCertStr, "\"passphrase\":\""+passphrase+"\"")
	}

	certStr = append(certStr, "\"rawCertificate\":{"+strings.Join(rawCertStr, ",")+"}")

	u.Path = path.Join(u.Path, "certificates")

	payload := "{" + strings.Join(certStr, ",") + "}"
	return apiclient.HttpClient(u.String(), payload)
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
	u.Path = path.Join(u.Path, "certificates")
	return apiclient.HttpClient(u.String())
}

// Delete
func Delete(name string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "certificates", name)
	return apiclient.HttpClient(u.String(), "", "DELETE")
}

// Get
func Get(name string) apiclient.APIResponse {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "certificates", name)
	return apiclient.HttpClient(u.String())
}

// Find
func Find(name string) (version string, err error) {
	cs := certs{}
	var respBody []byte

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())

	u.Path = path.Join(u.Path, "certificates")
	if response := apiclient.HttpClient(u.String()); response.Err != nil {
		return "", response.Err
	}

	if err = json.Unmarshal(respBody, &cs); err != nil {
		return "", apiclient.NewCliError("error unmarshalling", err)
	}

	for _, c := range cs.Cert {
		if c.DisplayName == name {
			version = c.Name[strings.LastIndex(c.Name, "/")+1:]
			return version, nil
		}
	}
	return "", fmt.Errorf("certificate not found")
}

// getStringifyiedContents
func getStringyfiedContents(file string) string {
	return strings.ReplaceAll(file, "\n", "\\n")
}
