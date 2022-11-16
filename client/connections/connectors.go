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

package connections

import (
	"encoding/json"
	"net/url"
	"path"
	"strconv"

	"github.com/srinandan/integrationcli/apiclient"
)

type connectionRequest struct {
	Name               string              `json:"name,omitempty"`
	Labels             map[string]string   `json:"labels,omitempty"`
	Description        string              `json:"description,omitempty"`
	ConnectorVersion   string              `json:"connectorVersion,omitempty"`
	ConfigVariables    []configVar         `json:"configVariables,omitempty"`
	LockConfig         lockConfig          `json:"lockConfig,omitempty"`
	DestinationConfigs []destinationConfig `json:"deatinationConfigs,omitempty"`
	AuthConfig         authConfig          `json:"authConfig,omitempty"`
	ServiceAccount     string              `json:"serviceAccount,omitempty"`
	Suspended          bool                `json:"suspended,omitempty"`
	NodeConfig         nodeConfig          `json:"nodeConfig,omitempty"`
}

type authConfig struct {
	AuthType                string                  `json:"authType,omitempty"`
	UserPassword            userPassword            `json:"userPassword,omitempty"`
	Oauth2JwtBearer         oauth2JwtBearer         `json:"oauth2JwtBearer,omitempty"`
	Oauth2ClientCredentials oauth2ClientCredentials `json:"oauth2ClientCredentials,omitempty"`
	SshPublicKey            sshPublicKey            `json:"sshPublicKey,omitempty"`
}

type lockConfig struct {
	Locked bool   `json:"locked,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type configVar struct {
	Key         string `json:"key,omitempty"`
	IntValue    string `json:"intValue,omitempty"`
	BoolValue   bool   `json:"boolValue,omitempty"`
	StringValue string `json:"stringValue,omitempty"`
	SecretValue secret `json:"secretValue,omitempty"`
}

type destinationConfig struct {
	Key          string        `json:"key,omitempty"`
	Destinations []destination `json:"destinations,omitempty"`
}

type userPassword struct {
	Username string `json:"username,omitempty"`
	Password secret `json:"password,omitempty"`
}

type oauth2JwtBearer struct {
	ClientKey secret    `json:"clientKey,omitempty"`
	JwtClaims jwtClaims `json:"jwtClaims,omitempty"`
}

type oauth2ClientCredentials struct {
	ClientId     string `json:"clientId,omitempty"`
	ClientSecret secret `json:"clientSecret,omitempty"`
}

type secret struct {
	SecretVersion string `json:"secretVersion,omitempty"`
}

type jwtClaims struct {
	Issuer   string `json:"issuer,omitempty"`
	Subject  string `json:"subject,omitempty"`
	Audience string `json:"audience,omitempty"`
}

type sshPublicKey struct {
	Username          string `json:"username,omitempty"`
	Password          secret `json:"password,omitempty"`
	SshClientCert     secret `json:"sshClientCert,omitempty"`
	CertType          string `json:"certType,omitempty"`
	SslClientCertPass secret `json:"sslClientCertPass,omitempty"`
}

type destination struct {
	Port              int    `json:"port,omitempty"`
	ServiceAttachment string `json:"serviceAttachment,omitempty"`
	Host              string `json:"host,omitempty"`
}

type nodeConfig struct {
	MinNodeCount string `json:"minNodeCount,omitempty"`
	MaxNodeCount string `json:"maxNodeCount,omitempty"`
}

// Create
func Create(name string, content []byte) (respBody []byte, err error) {
	c := connectionRequest{}
	if err = json.Unmarshal(content, &c); err != nil {
		return nil, err
	}

	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	q := u.Query()
	q.Set("connectionId", name)

	u.Path = path.Join(u.Path, name)
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), string(content))
	return respBody, err
}

// Delete
func Delete(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	u.Path = path.Join(u.Path, name)
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), "", "DELETE")
	return respBody, err
}

// Get
func Get(name string, view string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	q := u.Query()
	if view != "" {
		q.Set("view", view)
	}
	u.Path = path.Join(u.Path, name)
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	return respBody, err
}

// List
func List(pageSize int, pageToken string, filter string, orderBy string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
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
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	return respBody, err
}
