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
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/apigee/apigeecli/clilog"
	"github.com/srinandan/integrationcli/apiclient"
	"github.com/srinandan/integrationcli/cloudkms"
	"github.com/srinandan/integrationcli/secmgr"
)

const maxPageSize = 1000

type listconnections struct {
	Connections   []connection `json:"connections,omitempty"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

type connection struct {
	Name              *string           `json:"name,omitempty"`
	Description       string            `json:"description,omitempty"`
	ConnectorVersion  *string           `json:"connectorVersion,omitempty"`
	ConnectorDetails  *connectorDetails `json:"connectorDetails,omitempty"`
	ConfigVariables   []configVar       `json:"configVariables,omitempty"`
	AuthConfig        authConfig        `json:"authConfig,omitempty"`
	DestinationConfig destinationConfig `json:"destinationConfig,omitempty"`
	Suspended         bool              `json:"suspended,omitempty"`
}

type connectionRequest struct {
	Labels             *map[string]string   `json:"labels,omitempty"`
	Description        *string              `json:"description,omitempty"`
	ConnectorDetails   *connectorDetails    `json:"connectorDetails,omitempty"`
	ConnectorVersion   *string              `json:"connectorVersion,omitempty"`
	ConfigVariables    *[]configVar         `json:"configVariables,omitempty"`
	LockConfig         *lockConfig          `json:"lockConfig,omitempty"`
	DestinationConfigs *[]destinationConfig `json:"deatinationConfigs,omitempty"`
	AuthConfig         *authConfig          `json:"authConfig,omitempty"`
	ServiceAccount     *string              `json:"serviceAccount,omitempty"`
	Suspended          *bool                `json:"suspended,omitempty"`
	NodeConfig         *nodeConfig          `json:"nodeConfig,omitempty"`
}

type authConfig struct {
	AuthType                string                   `json:"authType,omitempty"`
	UserPassword            *userPassword            `json:"userPassword,omitempty"`
	Oauth2JwtBearer         *oauth2JwtBearer         `json:"oauth2JwtBearer,omitempty"`
	Oauth2ClientCredentials *oauth2ClientCredentials `json:"oauth2ClientCredentials,omitempty"`
	SshPublicKey            *sshPublicKey            `json:"sshPublicKey,omitempty"`
	AdditionalVariables     *[]configVar             `json:"additionalVariables,omitempty"`
}

type lockConfig struct {
	Locked bool   `json:"locked,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type connectorDetails struct {
	Name    string `json:"name,omitempty"`
	Version int    `json:"version,omitempty"`
}

type configVar struct {
	Key           string         `json:"key,omitempty"`
	IntValue      *string        `json:"intValue,omitempty"`
	BoolValue     *bool          `json:"boolValue,omitempty"`
	StringValue   *string        `json:"stringValue,omitempty"`
	SecretValue   *secret        `json:"secretValue,omitempty"`
	SecretDetails *secretDetails `json:"secretDetails,omitempty"`
}

type destinationConfig struct {
	Key          string        `json:"key,omitempty"`
	Destinations []destination `json:"destinations,omitempty"`
}

type userPassword struct {
	Username        string         `json:"username,omitempty"`
	Password        *secret        `json:"password,omitempty"`
	PasswordDetails *secretDetails `json:"passwordDetails,omitempty"`
}

type oauth2JwtBearer struct {
	ClientKey        *secret        `json:"clientKey,omitempty"`
	ClientKeyDetails *secretDetails `json:"clientKeyDetails,omitempty"`
	JwtClaims        jwtClaims      `json:"jwtClaims,omitempty"`
}

type oauth2ClientCredentials struct {
	ClientId            string         `json:"clientId,omitempty"`
	ClientSecret        *secret        `json:"clientSecret,omitempty"`
	ClientSecretDetails *secretDetails `json:"clientSecretDetails,omitempty"`
}

type secret struct {
	SecretVersion string `json:"secretVersion,omitempty"`
}

type secretDetails struct {
	SecretName string `json:"secretName,omitempty"`
	Reference  string `json:"reference,omitempty"`
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
	MinNodeCount int `json:"minNodeCount,omitempty"`
	MaxNodeCount int `json:"maxNodeCount,omitempty"`
}

// Create
func Create(name string, content []byte, serviceAccountName string, serviceAccountProject string, encryptionKey string, grantPermission bool) (respBody []byte, err error) {

	var secretVersion string

	c := connectionRequest{}
	if err = json.Unmarshal(content, &c); err != nil {
		return nil, err
	}

	//service account overrides have been provided, use them
	if serviceAccountName != "" {
		//set the project id if one was not presented
		if serviceAccountProject == "" {
			serviceAccountProject = apiclient.GetProjectID()
		}
		serviceAccountName = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", serviceAccountName, serviceAccountProject)
		//create the SA if it doesn't exist
		if err = apiclient.CreateServiceAccount(serviceAccountName); err != nil {
			return nil, err
		}
	} else if grantPermission { //use the default compute engine SA to grant permissions
		serviceAccountName, err = apiclient.GetComputeEngineDefaultServiceAccount(apiclient.GetProjectID())
		if err != nil {
			return nil, err
		}
	}

	if c.ServiceAccount == nil && serviceAccountName != "" {
		c.ServiceAccount = new(string)
		*c.ServiceAccount = serviceAccountName
	}

	if c.ConnectorDetails == nil {
		return nil, fmt.Errorf("connectorDetails must be set. See https://github.com/srinandan/integrationcli#connectors-for-third-party-applications for more details")
	}

	if c.ConnectorDetails.Name == "" || c.ConnectorDetails.Version < 0 {
		return nil, fmt.Errorf("connectorDetails Name and Version must be set. See https://github.com/srinandan/integrationcli#connectors-for-third-party-applications for more details")
	}

	//handle project id & region overrides
	if *c.ConfigVariables != nil && len(*c.ConfigVariables) > 0 {
		for index := range *c.ConfigVariables {
			if (*c.ConfigVariables)[index].Key == "project_id" && *(*c.ConfigVariables)[index].StringValue == "$PROJECT_ID$" {
				*(*c.ConfigVariables)[index].StringValue = apiclient.GetProjectID()
			} else if strings.Contains((*c.ConfigVariables)[index].Key, "_region") && *(*c.ConfigVariables)[index].StringValue == "$REGION$" {
				*(*c.ConfigVariables)[index].StringValue = apiclient.GetRegion()
			}
		}
	}

	// check if permissions need to be set
	if grantPermission && c.ServiceAccount != nil {
		var projectId string

		switch c.ConnectorDetails.Name {
		case "pubsub":
			var topicName string

			for _, configVar := range *c.ConfigVariables {
				if configVar.Key == "project_id" {
					projectId = *configVar.StringValue
				}
				if configVar.Key == "topic_id" {
					topicName = *configVar.StringValue
				}
			}

			if projectId == "" || topicName == "" {
				return nil, fmt.Errorf("projectId or topicName was not set")
			}

			if err = apiclient.SetPubSubIAMPermission(projectId, topicName, *c.ServiceAccount); err != nil {
				clilog.Warning.Printf("Unable to update permissions for the service account: %v\n", err)
			}
		case "bigquery":
			var datasetId string

			for _, configVar := range *c.ConfigVariables {
				if configVar.Key == "project_id" {
					projectId = *configVar.StringValue
				}
				if configVar.Key == "dataset_id" {
					datasetId = *configVar.StringValue
				}
			}
			if projectId == "" || datasetId == "" {
				return nil, fmt.Errorf("projectId or datasetId was not set")
			}

			if err = apiclient.SetBigQueryIAMPermission(projectId, datasetId, *c.ServiceAccount); err != nil {
				clilog.Warning.Printf("Unable to update permissions for the service account: %v\n", err)
			}
		case "gcs":
			for _, configVar := range *c.ConfigVariables {
				if configVar.Key == "project_id" {
					projectId = *configVar.StringValue
				}
			}
			if projectId == "" {
				return nil, fmt.Errorf("projectId was not set")
			}
			if err = apiclient.SetCloudStorageIAMPermission(projectId, *c.ServiceAccount); err != nil {
				clilog.Warning.Printf("Unable to update permissions for the service account: %v\n", err)
			}
		case "cloudsql-mysql", "cloudsql-postgresql", "cloudsql-sqlserver":
			for _, configVar := range *c.ConfigVariables {
				if configVar.Key == "project_id" {
					projectId = *configVar.StringValue
				}
			}
			if projectId == "" {
				return nil, fmt.Errorf("projectId was not set")
			}
			if err = apiclient.SetCloudSQLIAMPermission(projectId, *c.ServiceAccount); err != nil {
				clilog.Warning.Printf("Unable to update permissions for the service account: %v\n", err)
			}
		}
	}

	c.ConnectorVersion = new(string)
	*c.ConnectorVersion = fmt.Sprintf("projects/%s/locations/global/providers/gcp/connectors/%s/versions/%d",
		apiclient.GetProjectID(), c.ConnectorDetails.Name, c.ConnectorDetails.Version)

	//remove the element
	c.ConnectorDetails = nil

	//handle secrets for username
	if c.AuthConfig != nil && c.AuthConfig.UserPassword.PasswordDetails != nil {
		payload, err := readSecretFile(c.AuthConfig.UserPassword.PasswordDetails.Reference)
		if err != nil {
			return nil, err
		}

		//check if a Cloud KMS key was passsed, assume the file is encrypted
		if encryptionKey != "" {
			encryptionKey := path.Join("projects", apiclient.GetProjectID(), encryptionKey)
			payload, err = cloudkms.DecryptSymmetric(encryptionKey, payload)
			if err != nil {
				return nil, err
			}
		}

		if secretVersion, err = secmgr.Create(apiclient.GetProjectID(), c.AuthConfig.UserPassword.PasswordDetails.SecretName, payload); err != nil {
			return nil, err
		}
		secretName := c.AuthConfig.UserPassword.PasswordDetails.SecretName
		c.AuthConfig.UserPassword.Password = new(secret)
		c.AuthConfig.UserPassword.Password.SecretVersion = secretVersion
		c.AuthConfig.UserPassword.PasswordDetails = nil //clean the input

		if grantPermission && c.ServiceAccount != nil {
			//grant connector service account access to secretVersion
			if err = apiclient.SetSecretManagerIAMPermission(apiclient.GetProjectID(), secretName, *c.ServiceAccount); err != nil {
				return nil, err
			}
		}
	}

	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	q := u.Query()
	q.Set("connectionId", name)
	u.RawQuery = q.Encode()

	if content, err = json.Marshal(c); err != nil {
		return nil, err
	}

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

func Patch(name string, content []byte, updateMask []string) (respBody []byte, err error) {
	c := connectionRequest{}
	if err = json.Unmarshal(content, &c); err != nil {
		return nil, err
	}

	u, _ := url.Parse(apiclient.GetBaseConnectorURL())

	if len(updateMask) != 0 {
		updates := strings.Join(updateMask, ",")
		q := u.Query()
		q.Set("updateMask", updates)
		u.RawQuery = q.Encode()
	}

	u.Path = path.Join(u.Path, name)

	return apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), string(content), "PATCH")
}

func readSecretFile(name string) (payload []byte, err error) {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return nil, err
	}

	content, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// Import
func Import(folder string) (err error) {

	apiclient.SetPrintOutput(false)
	errs := []string{}

	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			clilog.Warning.Println("connection folder not found")
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(filepath.Base(path)))
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if _, err := Get(name, ""); err != nil { //create only if connection doesn't exist
			_, err = Create(name, content, "", "", "", false)
			if err != nil {
				errs = append(errs, err.Error())
			}
			fmt.Printf("creating connection %s\n", name)
		} else {
			fmt.Printf("connection %s already exists, skipping creations\n", name)
		}

		return nil
	})

	if err != nil {
		return nil
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

// Export
func Export(folder string) (err error) {
	apiclient.SetExportToFile(folder)
	apiclient.SetPrintOutput(false)

	respBody, err := List(maxPageSize, "", "", "")
	if err != nil {
		return err
	}

	lconnections := listconnections{}

	if err = json.Unmarshal(respBody, &lconnections); err != nil {
		return err
	}

	//no connections where found
	if len(lconnections.Connections) == 0 {
		return nil
	}

	for _, lconnection := range lconnections.Connections {
		lconnection.ConnectorDetails = new(connectorDetails)
		lconnection.ConnectorDetails.Name = getConnectorName(*lconnection.ConnectorVersion)
		lconnection.ConnectorDetails.Version = getConnectorVersion(*lconnection.ConnectorVersion)
		lconnection.ConnectorVersion = nil
		fileName := getConnectionName(*lconnection.Name) + ".json"
		lconnection.Name = nil
		connectionPayload, err := json.Marshal(lconnection)
		if err != nil {
			return err
		}
		if err = apiclient.WriteByteArrayToFile(path.Join(apiclient.GetExportToFile(), fileName), false, connectionPayload); err != nil {
			clilog.Error.Println(err)
			return err
		}
		fmt.Printf("Downloaded %s\n", fileName)
	}

	return nil
}

func getConnectorName(version string) string {
	return strings.Split(version, "/")[7]
}

func getConnectorVersion(version string) int {
	i, _ := strconv.Atoi(strings.Split(version, "/")[9])
	return i
}

func getConnectionName(name string) string {
	return name[strings.LastIndex(name, "/")+1:]
}
