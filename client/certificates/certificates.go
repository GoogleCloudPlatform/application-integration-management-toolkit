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
	"net/url"
	"path"
	"strconv"
	"strings"

	"internal/apiclient"
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
func Create(displayName string, description string, sslCertificate string, privateKey string, passphrase string) (respBody []byte, err error) {
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
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), payload)
	return respBody, err
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
	u.Path = path.Join(u.Path, "certificates")
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	return respBody, err
}

// Delete
func Delete(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "certificates", name)
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), "", "DELETE")
	return respBody, err
}

// Get
func Get(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	u.Path = path.Join(u.Path, "certificates", name)
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	return respBody, err
}

// Find
func Find(name string) (version string, err error) {
	cs := certs{}
	var respBody []byte

	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())

	u.Path = path.Join(u.Path, "certificates")
	if respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String()); err != nil {
		return "", err
	}

	if err = json.Unmarshal(respBody, &cs); err != nil {
		return "", err
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
