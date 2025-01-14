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
	"internal/apiclient"
	"internal/clilog"
	"internal/cloudkms"
	"internal/secmgr"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const maxPageSize = 1000

type listconnections struct {
	Connections   []connection `json:"connections,omitempty"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

type connection struct {
	Name                   *string             `json:"name,omitempty"`
	Description            string              `json:"description,omitempty"`
	ConnectorVersion       *string             `json:"connectorVersion,omitempty"`
	ConnectorDetails       *connectorDetails   `json:"connectorDetails,omitempty"`
	ConfigVariables        []configVar         `json:"configVariables,omitempty"`
	AuthConfig             authConfig          `json:"authConfig,omitempty"`
	NodeConfig             nodeConfig          `json:"nodeConfig,omitempty"`
	DestinationConfig      []destinationConfig `json:"destinationConfigs,omitempty"`
	Suspended              bool                `json:"suspended,omitempty"`
	LogConfig              *logConfig          `json:"logConfig,omitempty"`
	SslConfig              *sslConfig          `json:"sslConfig,omitempty"`
	EventingEnablementType *string             `json:"eventingEnablementType,omitempty"`
	EventingConfig         *eventingConfig     `json:"eventingConfig,omitempty"`
}

type connectionRequest struct {
	Labels                 *map[string]string   `json:"labels,omitempty"`
	Description            *string              `json:"description,omitempty"`
	ConnectorDetails       *connectorDetails    `json:"connectorDetails,omitempty"`
	ConnectorVersion       *string              `json:"connectorVersion,omitempty"`
	ConfigVariables        *[]configVar         `json:"configVariables,omitempty"`
	LockConfig             *lockConfig          `json:"lockConfig,omitempty"`
	DestinationConfigs     *[]destinationConfig `json:"destinationConfigs,omitempty"`
	AuthConfig             *authConfig          `json:"authConfig,omitempty"`
	ServiceAccount         *string              `json:"serviceAccount,omitempty"`
	Suspended              *bool                `json:"suspended,omitempty"`
	NodeConfig             *nodeConfig          `json:"nodeConfig,omitempty"`
	LogConfig              *logConfig           `json:"logConfig,omitempty"`
	SslConfig              *sslConfig           `json:"sslConfig,omitempty"`
	EventingEnablementType *string              `json:"eventingEnablementType,omitempty"`
	EventingConfig         *eventingConfig      `json:"eventingConfig,omitempty"`
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

type logConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

type sslConfig struct {
	UseSSL                   bool                      `json:"useSsl,omitempty"`
	Type                     *string                   `json:"type,omitempty"`
	PrivateServerCertificate *privateServerCertificate `json:"privateServerCertificate,omitempty"`
	ClientCertificate        *clientCertificate        `json:"clientCertificate,omitempty"`
	ClientPrivateKey         *clientPrivateKey         `json:"clientPrivateKey,omitempty"`
	ClientPrivateKeyPass     *clientPrivateKeyPass     `json:"clientPrivateKeyPass,omitempty"`
	ClientCertType           *string                   `json:"clientCertType,omitempty"`
	ServerCertType           *string                   `json:"serverCertType,omitempty"`
}

type connectorDetails struct {
	Name      string  `json:"name,omitempty"`
	Provider  string  `json:"provider,omitempty"`
	Version   *int    `json:"version,omitempty"`
	VersionId *string `json:"versionId,omitempty"`
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
	Username                 string         `json:"username,omitempty"`
	Password                 *secret        `json:"password,omitempty"`
	PasswordDetails          *secretDetails `json:"passwordDetails,omitempty"`
	SshClientCert            *secret        `json:"sshClientCert,omitempty"`
	SshClientCertDetails     *secretDetails `json:"sshClientCertDetails,omitempty"`
	CertType                 string         `json:"certType,omitempty"`
	SslClientCertPass        *secret        `json:"sslClientCertPass,omitempty"`
	SslClientCertPassDetails *secretDetails `json:"sslClientCertPassDetails,omitempty"`
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

type privateServerCertificate struct {
	SecretVersion *string        `json:"secretVersion,omitempty"`
	SecretDetails *secretDetails `json:"secretDetails,omitempty"`
}

type clientCertificate struct {
	SecretVersion *string        `json:"secretVersion,omitempty"`
	SecretDetails *secretDetails `json:"secretDetails,omitempty"`
}

type clientPrivateKey struct {
	SecretVersion *string        `json:"secretVersion,omitempty"`
	SecretDetails *secretDetails `json:"secretDetails,omitempty"`
}

type clientPrivateKeyPass struct {
	SecretVersion *string        `json:"secretVersion,omitempty"`
	SecretDetails *secretDetails `json:"secretDetails,omitempty"`
}

type status struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type operation struct {
	Name     string                  `json:"name,omitempty"`
	Done     bool                    `json:"done,omitempty"`
	Error    *status                 `json:"error,omitempty"`
	Response *map[string]interface{} `json:"response,omitempty"`
}

type eventingConfig struct {
	EnrichmentEnabled             bool                  `json:"enrichmentEnabled,omitempty"`
	PrivateConnectivityEnabled    bool                  `json:"privateConnectivityEnabled,omitempty"`
	EventsListenerIngressEndpoint string                `json:"eventsListenerIngressEndpoint,omitempty"`
	AdditionalVariables           []additionalVariables `json:"additionalVariables,omitempty"`
	RegistrationDestinationConfig *destinationConfig    `json:"registrationDestinationConfig,omitempty"`
	AuthConfig                    *authConfig           `json:"authConfig,omitempty"`
	ListenerAuthConfig            authConfig            `json:"listenerAuthConfig,omitempty"`
	DeadLetterConfig              deadLetterConfig      `json:"deadLetterConfig,omitempty"`
	ProxyDestinationConfig        *destinationConfig    `json:"proxyDestinationConfig,omitempty"`
}

type additionalVariables struct {
	Key                string           `json:"key,omitempty"`
	IntValue           string           `json:"intValue,omitempty"`
	BoolValue          bool             `json:"boolValue,omitempty"`
	StringValue        string           `json:"stringValue,omitempty"`
	SecretValue        *secretValue     `json:"secretValue,omitempty"`
	EncryptionKeyValue *encryptionValue `json:"encryptionKeyValue,omitempty"`
}

type secretValue struct {
	SecretVersion *string `json:"secretVersion,omitempty"`
}

type encryptionValue struct {
	Type       string `json:"type,omitempty"`
	KmsKeyName string `json:"kmsKeyName,omitempty"`
}

type deadLetterConfig struct {
	Topic     string `json:"topic,omitempty"`
	ProjectId string `json:"projectId,omitempty"`
}

type proxyDestinationConfig struct {
	Destinations []destination `json:"destinations,omitempty"`
}

const interval = 10

// Create
func Create(name string, content []byte, serviceAccountName string, serviceAccountProject string,
	encryptionKey string, grantPermission bool, createSecret bool, wait bool,
) (respBody []byte, err error) {
	if serviceAccountName != "" && strings.Contains(serviceAccountName, ".iam.gserviceaccount.com") {
		serviceAccountName = strings.Split(serviceAccountName, "@")[0]
	}

	operationsBytes, err := create(name, content, serviceAccountName,
		serviceAccountProject, encryptionKey, grantPermission, createSecret)
	if err != nil {
		return nil, err
	}

	if wait {
		apiclient.ClientPrintHttpResponse.Set(false)
		defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

		o := operation{}
		if err = json.Unmarshal(operationsBytes, &o); err != nil {
			return nil, err
		}

		operationId := filepath.Base(o.Name)
		clilog.Info.Printf("Checking connection status for %s in %d seconds\n", operationId, interval)

		stop := apiclient.Every(interval*time.Second, func(time.Time) bool {
			var respBody []byte

			if respBody, err = GetOperation(operationId); err != nil {
				return false
			}

			if err = json.Unmarshal(respBody, &o); err != nil {
				return false
			}

			if o.Done {
				if o.Error != nil {
					clilog.Error.Printf("Connection completed with error: %s\n", o.Error.Message)
				} else {
					clilog.Info.Println("Connection completed successfully!")
				}
				return false
			} else {
				clilog.Info.Printf("Connection status is: %t. Waiting %d seconds.\n", o.Done, interval)
				return true
			}
		})

		<-stop
	}

	return respBody, err
}

// create
func create(name string, content []byte, serviceAccountName string, serviceAccountProject string,
	encryptionKey string, grantPermission bool, createSecret bool,
) (respBody []byte, err error) {
	var secretVersion string

	c := connectionRequest{}
	if err = json.Unmarshal(content, &c); err != nil {
		return nil, err
	}

	// service account overrides have been provided, use them
	if serviceAccountName != "" {
		// set the project id if one was not presented
		if serviceAccountProject == "" {
			serviceAccountProject = apiclient.GetProjectID()
		}
		serviceAccountName = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", serviceAccountName, serviceAccountProject)
		// create the SA if it doesn't exist
		if err = apiclient.CreateServiceAccount(serviceAccountName); err != nil {
			return nil, err
		}
	} else if grantPermission { // use the default compute engine SA to grant permissions
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
		return nil, fmt.Errorf("connectorDetails must be set." +
			" See https://github.com/GoogleCloudPlatform/application-integration-management-toolkit" +
			"#connectors-for-third-party-applications for more details")
	}

	if c.ConnectorDetails.Version != nil && c.ConnectorDetails.VersionId != nil {
		return nil, fmt.Errorf("Version and VersionId cannot be set")
	}

	if c.ConnectorDetails.Name == "" || c.ConnectorDetails.Provider == "" {
		return nil, fmt.Errorf("connectorDetails Name and Provider must be set." +
			" See https://github.com/GoogleCloudPlatform/application-integration-management-toolkit" +
			"#connectors-for-third-party-applications for more details")
	}

	if c.ConnectorDetails.Provider == "customconnector" && c.ConnectorDetails.VersionId == nil {
		return nil, fmt.Errorf("connectorDetails VersionId must be set for customconnectors")
	} else if c.ConnectorDetails.Provider != "customconnector" && c.ConnectorDetails.Version == nil {
		return nil, fmt.Errorf("connectorDetails Version must be set")
	}

	// handle project id & region overrides
	if c.ConfigVariables != nil && len(*c.ConfigVariables) > 0 {
		for index := range *c.ConfigVariables {
			if (*c.ConfigVariables)[index].Key == "project_id" && *(*c.ConfigVariables)[index].StringValue == "$PROJECT_ID$" {
				*(*c.ConfigVariables)[index].StringValue = apiclient.GetProjectID()
			} else if strings.Contains((*c.ConfigVariables)[index].Key, "_region") &&
				*(*c.ConfigVariables)[index].StringValue == "$REGION$" {
				*(*c.ConfigVariables)[index].StringValue = apiclient.GetRegion()
			}
		}
	}

	// check if permissions need to be set
	if grantPermission && c.ServiceAccount != nil {
		var projectID string

		switch c.ConnectorDetails.Name {
		case "pubsub":
			var topicName string

			for _, configVar := range *c.ConfigVariables {
				if configVar.Key == "project_id" {
					projectID = *configVar.StringValue
				}
				if configVar.Key == "topic_id" {
					topicName = *configVar.StringValue
				}
			}

			if projectID == "" || topicName == "" {
				return nil, fmt.Errorf("projectId or topicName was not set")
			}

			if err = apiclient.SetPubSubIAMPermission(projectID, topicName, *c.ServiceAccount); err != nil {
				clilog.Warning.Printf("Unable to update permissions for the service account: %v\n", err)
			}
		case "bigquery":
			var datasetID string

			for _, configVar := range *c.ConfigVariables {
				if configVar.Key == "project_id" {
					projectID = *configVar.StringValue
				}
				if configVar.Key == "dataset_id" {
					datasetID = *configVar.StringValue
				}
			}
			if projectID == "" || datasetID == "" {
				return nil, fmt.Errorf("project_id or dataset_id was not set")
			}

			if err = apiclient.SetBigQueryIAMPermission(projectID, datasetID, *c.ServiceAccount); err != nil {
				clilog.Warning.Printf("Unable to update permissions for the service account: %v\n", err)
			}
		case "gcs":
			for _, configVar := range *c.ConfigVariables {
				if configVar.Key == "project_id" {
					projectID = *configVar.StringValue
				}
			}
			if projectID == "" {
				return nil, fmt.Errorf("project_id was not set")
			}
			if err = apiclient.SetCloudStorageIAMPermission(projectID, *c.ServiceAccount); err != nil {
				clilog.Warning.Printf("Unable to update permissions for the service account: %v\n", err)
			}
		case "cloudsql-mysql", "cloudsql-postgresql", "cloudsql-sqlserver":
			for _, configVar := range *c.ConfigVariables {
				if configVar.Key == "project_id" {
					projectID = *configVar.StringValue
				}
			}
			if projectID == "" {
				return nil, fmt.Errorf("projectId was not set")
			}
			if err = apiclient.SetCloudSQLIAMPermission(projectID, *c.ServiceAccount); err != nil {
				clilog.Warning.Printf("Unable to update permissions for the service account: %v\n", err)
			}
		case "cloudspanner":
			for _, configVar := range *c.ConfigVariables {
				if configVar.Key == "project_id" {
					projectID = *configVar.StringValue
				}
			}
			if projectID == "" {
				return nil, fmt.Errorf("project_id was not set")
			}
			if err = apiclient.SetCloudSpannerIAMPermission(projectID, *c.ServiceAccount); err != nil {
				clilog.Warning.Printf("Unable to update permissions for the service account: %v\n", err)
			}
		}
	}

	c.ConnectorVersion = new(string)
	if c.ConnectorDetails.VersionId != nil {
		*c.ConnectorVersion = fmt.Sprintf("projects/%s/locations/global/providers/%s/connectors/%s/versions/%s",
			apiclient.GetProjectID(), c.ConnectorDetails.Provider, c.ConnectorDetails.Name, *c.ConnectorDetails.VersionId)
	} else {
		*c.ConnectorVersion = fmt.Sprintf("projects/%s/locations/global/providers/%s/connectors/%s/versions/%d",
			apiclient.GetProjectID(), c.ConnectorDetails.Provider, c.ConnectorDetails.Name, *c.ConnectorDetails.Version)
	}

	// remove the element
	c.ConnectorDetails = nil

	// handle secrets for username
	if c.AuthConfig != nil {
		switch c.AuthConfig.AuthType {
		case "USER_PASSWORD":
			if c.AuthConfig.UserPassword != nil && c.AuthConfig.UserPassword.PasswordDetails != nil {
				if createSecret {
					if c.AuthConfig.UserPassword.PasswordDetails.Reference == "" {
						return nil, fmt.Errorf("create-secret is enabled, but reference is not passed")
					}
					payload, err := readSecretFile(c.AuthConfig.UserPassword.PasswordDetails.Reference)
					if err != nil {
						return nil, err
					}

					// check if a Cloud KMS key was passsed, assume the file is encrypted
					if encryptionKey != "" {
						encryptionKey := path.Join("projects", apiclient.GetProjectID(), encryptionKey)
						payload, err = cloudkms.DecryptSymmetric(encryptionKey, payload)
						if err != nil {
							return nil, err
						}
					}

					if secretVersion, err = secmgr.Create(
						apiclient.GetProjectID(),
						c.AuthConfig.UserPassword.PasswordDetails.SecretName,
						payload); err != nil {
						return nil, err
					}

					secretName := c.AuthConfig.UserPassword.PasswordDetails.SecretName
					c.AuthConfig.UserPassword.Password = new(secret)
					c.AuthConfig.UserPassword.Password.SecretVersion = secretVersion
					c.AuthConfig.UserPassword.PasswordDetails = nil // clean the input
					if grantPermission && c.ServiceAccount != nil {
						// grant connector service account access to secretVersion
						if err = apiclient.SetSecretManagerIAMPermission(
							apiclient.GetProjectID(),
							secretName,
							*c.ServiceAccount); err != nil {
							return nil, err
						}
					}
				} else {
					c.AuthConfig.UserPassword.Password = new(secret)
					c.AuthConfig.UserPassword.Password.SecretVersion = fmt.Sprintf("projects/%s/secrets/%s/versions/1",
						apiclient.GetProjectID(), c.AuthConfig.UserPassword.PasswordDetails.SecretName)
					c.AuthConfig.UserPassword.PasswordDetails = nil // clean the input
				}
			}
		case "OAUTH2_JWT_BEARER":
			if c.AuthConfig.Oauth2JwtBearer != nil && c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails != nil {
				if createSecret {
					clilog.Warning.Printf("Creating secrets for %s is not implemented\n", c.AuthConfig.AuthType)
					payload, err := readSecretFile(c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails.Reference)
					if err != nil {
						return nil, err
					}
					// check if a Cloud KMS key was passsed, assume the file is encrypted
					if encryptionKey != "" {
						encryptionKey := path.Join("projects", apiclient.GetProjectID(), encryptionKey)
						payload, err = cloudkms.DecryptSymmetric(encryptionKey, payload)
						if err != nil {
							return nil, err
						}
					}
					if secretVersion, err = secmgr.Create(
						apiclient.GetProjectID(),
						c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails.SecretName,
						payload); err != nil {
						return nil, err
					}
					secretName := c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails.SecretName
					c.AuthConfig.Oauth2JwtBearer.ClientKey = new(secret)
					c.AuthConfig.Oauth2JwtBearer.ClientKey.SecretVersion = secretVersion
					c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails = nil // clean the input
					if grantPermission && c.ServiceAccount != nil {
						// grant connector service account access to secret version
						if err = apiclient.SetSecretManagerIAMPermission(
							apiclient.GetProjectID(),
							secretName,
							*c.ServiceAccount); err != nil {
							return nil, err
						}
					}
				} else {
					c.AuthConfig.Oauth2JwtBearer.ClientKey = new(secret)
					c.AuthConfig.Oauth2JwtBearer.ClientKey.SecretVersion = fmt.Sprintf("projects/%s/secrets/%s/versions/1",
						apiclient.GetProjectID(),
						c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails.SecretName)
					c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails = nil
				}
			}
		case "OAUTH2_CLIENT_CREDENTIALS":
			if createSecret {
				clilog.Warning.Printf("Creating secrets for %s is not implemented\n", c.AuthConfig.AuthType)
			}
		case "SSH_PUBLIC_KEY":
			if createSecret {
				clilog.Warning.Printf("Creating secrets for %s is not implemented\n", c.AuthConfig.AuthType)
			}
		case "OAUTH2_AUTH_CODE_FLOW":
			if createSecret {
				clilog.Warning.Printf("Creating secrets for %s is not implemented\n", c.AuthConfig.AuthType)
			}
		default:
			clilog.Warning.Printf("No auth type found, assuming service account auth\n")
		}
	}

	// handle secrets for ssl config
	if c.SslConfig != nil {
		if c.SslConfig.PrivateServerCertificate != nil && c.SslConfig.PrivateServerCertificate.SecretDetails != nil {
			if createSecret {
				payload, err := readSecretFile(c.SslConfig.PrivateServerCertificate.SecretDetails.Reference)
				if err != nil {
					return nil, err
				}
				// check if a Cloud KMS key was passsed, assume the file is encrypted
				if encryptionKey != "" {
					encryptionKey := path.Join("projects", apiclient.GetProjectID(), encryptionKey)
					payload, err = cloudkms.DecryptSymmetric(encryptionKey, payload)
					if err != nil {
						return nil, err
					}
				}

				if secretVersion, err = secmgr.Create(
					apiclient.GetProjectID(),
					c.SslConfig.PrivateServerCertificate.SecretDetails.SecretName,
					payload); err != nil {
					return nil, err
				}

				c.SslConfig.PrivateServerCertificate.SecretVersion = new(string)
				*c.SslConfig.PrivateServerCertificate.SecretVersion = secretVersion
				c.SslConfig.PrivateServerCertificate.SecretDetails = nil // clean the input

			} else {
				c.SslConfig.PrivateServerCertificate.SecretVersion = new(string)
				*c.SslConfig.PrivateServerCertificate.SecretVersion = fmt.Sprintf("projects/%s/secrets/%s/versions/1",
					apiclient.GetProjectID(), c.SslConfig.PrivateServerCertificate.SecretDetails.SecretName)
				c.SslConfig.PrivateServerCertificate.SecretDetails = nil // clean the input
			}
		}
		if c.SslConfig.ClientCertificate != nil && c.SslConfig.ClientCertificate.SecretDetails != nil {
			if createSecret {
				payload, err := readSecretFile(c.SslConfig.ClientCertificate.SecretDetails.Reference)
				if err != nil {
					return nil, err
				}
				// check if a Cloud KMS key was passsed, assume the file is encrypted
				if encryptionKey != "" {
					encryptionKey := path.Join("projects", apiclient.GetProjectID(), encryptionKey)
					payload, err = cloudkms.DecryptSymmetric(encryptionKey, payload)
					if err != nil {
						return nil, err
					}
				}

				if secretVersion, err = secmgr.Create(
					apiclient.GetProjectID(),
					c.SslConfig.ClientCertificate.SecretDetails.SecretName,
					payload); err != nil {
					return nil, err
				}

				c.SslConfig.ClientCertificate.SecretVersion = new(string)
				*c.SslConfig.ClientCertificate.SecretVersion = secretVersion
				c.SslConfig.ClientCertificate.SecretDetails = nil // clean the input
			} else {
				c.SslConfig.ClientCertificate.SecretVersion = new(string)
				*c.SslConfig.ClientCertificate.SecretVersion = fmt.Sprintf("projects/%s/secrets/%s/versions/1",
					apiclient.GetProjectID(), c.SslConfig.ClientCertificate.SecretDetails.SecretName)
				c.SslConfig.ClientCertificate.SecretDetails = nil // clean the input
			}
		}
		if c.SslConfig.ClientPrivateKey != nil && c.SslConfig.ClientPrivateKey.SecretDetails != nil {
			if createSecret {
				payload, err := readSecretFile(c.SslConfig.ClientPrivateKey.SecretDetails.Reference)
				if err != nil {
					return nil, err
				}
				// check if a Cloud KMS key was passsed, assume the file is encrypted
				if encryptionKey != "" {
					encryptionKey := path.Join("projects", apiclient.GetProjectID(), encryptionKey)
					payload, err = cloudkms.DecryptSymmetric(encryptionKey, payload)
					if err != nil {
						return nil, err
					}
				}

				if secretVersion, err = secmgr.Create(
					apiclient.GetProjectID(),
					c.SslConfig.ClientPrivateKey.SecretDetails.SecretName,
					payload); err != nil {
					return nil, err
				}

				c.SslConfig.ClientPrivateKey.SecretVersion = new(string)
				*c.SslConfig.ClientPrivateKey.SecretVersion = secretVersion
				c.SslConfig.ClientPrivateKey.SecretDetails = nil // clean the input
			} else {
				c.SslConfig.ClientPrivateKey.SecretVersion = new(string)
				*c.SslConfig.ClientPrivateKey.SecretVersion = fmt.Sprintf("projects/%s/secrets/%s/versions/1",
					apiclient.GetProjectID(), c.SslConfig.ClientPrivateKey.SecretDetails.SecretName)
				c.SslConfig.ClientPrivateKey.SecretDetails = nil // clean the input
			}
		}
		if c.SslConfig.ClientPrivateKeyPass != nil && c.SslConfig.ClientPrivateKeyPass.SecretDetails != nil {
			if createSecret {
				payload, err := readSecretFile(c.SslConfig.ClientPrivateKeyPass.SecretDetails.Reference)
				if err != nil {
					return nil, err
				}
				// check if a Cloud KMS key was passsed, assume the file is encrypted
				if encryptionKey != "" {
					encryptionKey := path.Join("projects", apiclient.GetProjectID(), encryptionKey)
					payload, err = cloudkms.DecryptSymmetric(encryptionKey, payload)
					if err != nil {
						return nil, err
					}
				}

				if secretVersion, err = secmgr.Create(
					apiclient.GetProjectID(),
					c.SslConfig.ClientPrivateKeyPass.SecretDetails.SecretName,
					payload); err != nil {
					return nil, err
				}

				c.SslConfig.ClientPrivateKeyPass.SecretVersion = new(string)
				*c.SslConfig.ClientPrivateKeyPass.SecretVersion = secretVersion
				c.SslConfig.ClientPrivateKeyPass.SecretDetails = nil // clean the input
			} else {
				c.SslConfig.ClientPrivateKeyPass.SecretVersion = new(string)
				*c.SslConfig.ClientPrivateKeyPass.SecretVersion = fmt.Sprintf("projects/%s/secrets/%s/versions/1",
					apiclient.GetProjectID(), c.SslConfig.ClientPrivateKeyPass.SecretDetails.SecretName)
				c.SslConfig.ClientPrivateKeyPass.SecretDetails = nil // clean the input
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

	respBody, err = apiclient.HttpClient(u.String(), string(content))
	return respBody, err
}

// Delete
func Delete(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	u.Path = path.Join(u.Path, name)
	respBody, err = apiclient.HttpClient(u.String(), "", "DELETE")
	return respBody, err
}

// Get
func Get(name string, view string, minimal bool, overrides bool) (respBody []byte, err error) {
	var connectionPayload []byte
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	q := u.Query()
	if view != "" {
		q.Set("view", view)
	}
	u.Path = path.Join(u.Path, name)

	if minimal {
		apiclient.ClientPrintHttpResponse.Set(false)
	}

	respBody, err = apiclient.HttpClient(u.String())

	if minimal {
		c := connection{}
		err := json.Unmarshal(respBody, &c)
		if err != nil {
			return nil, err
		}

		c.ConnectorDetails = new(connectorDetails)
		c.ConnectorDetails.Name = getConnectorName(*c.ConnectorVersion)
		c.ConnectorDetails.Provider = getConnectorProvider(*c.ConnectorVersion)
		if c.ConnectorDetails.Provider != "customconnector" {
			c.ConnectorDetails.Version = new(int)
			*c.ConnectorDetails.Version = getConnectorVersion(*c.ConnectorVersion)
		} else {
			c.ConnectorDetails.VersionId = new(string)
			*c.ConnectorDetails.VersionId = getConnectorVersionId(*c.ConnectorVersion)
		}

		c.ConnectorVersion = nil
		c.Name = nil
		if overrides {
			switch c.AuthConfig.AuthType {
			case "USER_PASSWORD":
				p := c.AuthConfig.UserPassword.Password.SecretVersion
				c.AuthConfig.UserPassword.PasswordDetails = new(secretDetails)
				c.AuthConfig.UserPassword.PasswordDetails.SecretName = strings.Split(p, "/")[3]
				c.AuthConfig.UserPassword.Password = nil
			case "OAUTH2_JWT_BEARER":
				p := c.AuthConfig.Oauth2JwtBearer.ClientKey.SecretVersion
				c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails = new(secretDetails)
				c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails.SecretName = strings.Split(p, "/")[3]
				c.AuthConfig.Oauth2JwtBearer.ClientKey = nil
			}
			if isGoogleConnection(c.ConnectorDetails.Name) {
				for _, configVar := range c.ConfigVariables {
					if configVar.Key == "project_id" {
						*configVar.StringValue = "$PROJECT_ID$"
					}
				}
			}
			if c.SslConfig != nil {
				if c.SslConfig.PrivateServerCertificate != nil && c.SslConfig.PrivateServerCertificate.SecretVersion != nil {
					p := *c.SslConfig.PrivateServerCertificate.SecretVersion
					c.SslConfig.PrivateServerCertificate.SecretDetails = new(secretDetails)
					c.SslConfig.PrivateServerCertificate.SecretDetails.SecretName = strings.Split(p, "/")[3]
					c.SslConfig.PrivateServerCertificate.SecretVersion = nil
				}
				if c.SslConfig.ClientCertificate != nil && c.SslConfig.ClientCertificate.SecretVersion != nil {
					p := *c.SslConfig.ClientCertificate.SecretVersion
					c.SslConfig.ClientCertificate.SecretDetails = new(secretDetails)
					c.SslConfig.ClientCertificate.SecretDetails.SecretName = strings.Split(p, "/")[3]
					c.SslConfig.ClientCertificate.SecretVersion = nil
				}
				if c.SslConfig.ClientPrivateKey != nil && c.SslConfig.ClientPrivateKey.SecretVersion != nil {
					p := *c.SslConfig.ClientPrivateKey.SecretVersion
					c.SslConfig.ClientPrivateKey.SecretDetails = new(secretDetails)
					c.SslConfig.ClientPrivateKey.SecretDetails.SecretName = strings.Split(p, "/")[3]
					c.SslConfig.ClientPrivateKey.SecretVersion = nil
				}
				if c.SslConfig.ClientPrivateKeyPass != nil && c.SslConfig.ClientPrivateKeyPass.SecretVersion != nil {
					p := *c.SslConfig.ClientPrivateKeyPass.SecretVersion
					c.SslConfig.ClientPrivateKeyPass.SecretDetails = new(secretDetails)
					c.SslConfig.ClientPrivateKeyPass.SecretDetails.SecretName = strings.Split(p, "/")[3]
					c.SslConfig.ClientPrivateKeyPass.SecretVersion = nil
				}
			}
		}
		connectionPayload, err = json.Marshal(c)
		if err != nil {
			return nil, err
		}
		apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting()) // set original print output
		apiclient.PrettyPrint(connectionPayload)

		return connectionPayload, err
	}
	return respBody, err
}

// Get Connection details With region
func GetConnectionDetailWithRegion(name string, region string, view string, minimal bool, overrides bool) (respBody []byte, err error) {
	var connectionPayload []byte
	u, _ := url.Parse(apiclient.GetBaseConnectorURLWithRegion(region))
	q := u.Query()
	if view != "" {
		q.Set("view", view)
	}
	u.Path = path.Join(u.Path, name)

	if minimal {
		apiclient.ClientPrintHttpResponse.Set(false)
	}

	respBody, err = apiclient.HttpClient(u.String())

	if minimal {
		c := connection{}
		err := json.Unmarshal(respBody, &c)
		if err != nil {
			return nil, err
		}

		c.ConnectorDetails = new(connectorDetails)
		c.ConnectorDetails.Name = getConnectorName(*c.ConnectorVersion)
		c.ConnectorDetails.Provider = getConnectorProvider(*c.ConnectorVersion)
		if c.ConnectorDetails.Provider != "customconnector" {
			c.ConnectorDetails.Version = new(int)
			*c.ConnectorDetails.Version = getConnectorVersion(*c.ConnectorVersion)
		} else {
			c.ConnectorDetails.VersionId = new(string)
			*c.ConnectorDetails.VersionId = getConnectorVersionId(*c.ConnectorVersion)
		}

		c.ConnectorVersion = nil
		c.Name = nil
		if overrides {
			switch c.AuthConfig.AuthType {
			case "USER_PASSWORD":
				if c.AuthConfig.UserPassword != nil && c.AuthConfig.UserPassword.Password != nil {
					p := c.AuthConfig.UserPassword.Password.SecretVersion
					c.AuthConfig.UserPassword.PasswordDetails = new(secretDetails)
					c.AuthConfig.UserPassword.PasswordDetails.SecretName = strings.Split(p, "/")[3]
					c.AuthConfig.UserPassword.Password = nil
				}
			case "OAUTH2_JWT_BEARER":
				if c.AuthConfig.Oauth2JwtBearer != nil && c.AuthConfig.Oauth2JwtBearer.ClientKey != nil {
					p := c.AuthConfig.Oauth2JwtBearer.ClientKey.SecretVersion
					c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails = new(secretDetails)
					c.AuthConfig.Oauth2JwtBearer.ClientKeyDetails.SecretName = strings.Split(p, "/")[3]
					c.AuthConfig.Oauth2JwtBearer.ClientKey = nil
				}
			}
			if isGoogleConnection(c.ConnectorDetails.Name) {
				for _, configVar := range c.ConfigVariables {
					if configVar.Key == "project_id" {
						*configVar.StringValue = "$PROJECT_ID$"
					}
				}
			}
			if c.SslConfig != nil {
				if c.SslConfig.PrivateServerCertificate != nil && c.SslConfig.PrivateServerCertificate.SecretVersion != nil {
					p := *c.SslConfig.PrivateServerCertificate.SecretVersion
					c.SslConfig.PrivateServerCertificate.SecretDetails = new(secretDetails)
					c.SslConfig.PrivateServerCertificate.SecretDetails.SecretName = strings.Split(p, "/")[3]
					c.SslConfig.PrivateServerCertificate.SecretVersion = nil
				}
				if c.SslConfig.ClientCertificate != nil && c.SslConfig.ClientCertificate.SecretVersion != nil {
					p := *c.SslConfig.ClientCertificate.SecretVersion
					c.SslConfig.ClientCertificate.SecretDetails = new(secretDetails)
					c.SslConfig.ClientCertificate.SecretDetails.SecretName = strings.Split(p, "/")[3]
					c.SslConfig.ClientCertificate.SecretVersion = nil
				}
				if c.SslConfig.ClientPrivateKey != nil && c.SslConfig.ClientPrivateKey.SecretVersion != nil {
					p := *c.SslConfig.ClientPrivateKey.SecretVersion
					c.SslConfig.ClientPrivateKey.SecretDetails = new(secretDetails)
					c.SslConfig.ClientPrivateKey.SecretDetails.SecretName = strings.Split(p, "/")[3]
					c.SslConfig.ClientPrivateKey.SecretVersion = nil
				}
				if c.SslConfig.ClientPrivateKeyPass != nil && c.SslConfig.ClientPrivateKeyPass.SecretVersion != nil {
					p := *c.SslConfig.ClientPrivateKeyPass.SecretVersion
					c.SslConfig.ClientPrivateKeyPass.SecretDetails = new(secretDetails)
					c.SslConfig.ClientPrivateKeyPass.SecretDetails.SecretName = strings.Split(p, "/")[3]
					c.SslConfig.ClientPrivateKeyPass.SecretVersion = nil
				}
			}
		}
		connectionPayload, err = json.Marshal(c)
		if err != nil {
			return nil, err
		}
		apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting()) // set original print output
		apiclient.PrettyPrint(connectionPayload)

		return connectionPayload, err
	}
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
	respBody, err = apiclient.HttpClient(u.String())
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

	return apiclient.HttpClient(u.String(), string(content), "PATCH")
}

func readSecretFile(name string) (payload []byte, err error) {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to open secret file %s, err: %w", name, err)
	}

	content, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// Import
func Import(folder string, createSecret bool, wait bool) (err error) {
	apiclient.ClientPrintHttpResponse.Set(false)
	defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())
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
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if _, err := Get(name, "", false, false); err != nil { // create only if connection doesn't exist
			_, err = Create(name, content, "", "", "", false, createSecret, wait)
			if err != nil {
				errs = append(errs, err.Error())
			}
			clilog.Info.Printf("creating connection %s\n", name)
		} else {
			clilog.Info.Printf("connection %s already exists, skipping creations\n", name)
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
	apiclient.ClientPrintHttpResponse.Set(false)
	defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

	pageToken := ""
	lconnections := listconnections{}

	for {
		l := listconnections{}
		respBody, err := List(maxPageSize, pageToken, "", "")
		if err != nil {
			return fmt.Errorf("failed to fetch Integrations: %w", err)
		}
		err = json.Unmarshal(respBody, &l)
		if err != nil {
			return fmt.Errorf("failed to unmarshall: %w", err)
		}
		lconnections.Connections = append(lconnections.Connections, l.Connections...)
		pageToken = l.NextPageToken
		if l.NextPageToken == "" {
			break
		}
	}

	// no connections where found
	if len(lconnections.Connections) == 0 {
		return nil
	}

	for _, lconnection := range lconnections.Connections {
		lconnection.ConnectorDetails = new(connectorDetails)
		lconnection.ConnectorDetails.Name = getConnectorName(*lconnection.ConnectorVersion)
		if lconnection.ConnectorDetails.Provider != "customconnector" {
			lconnection.ConnectorDetails.Version = new(int)
			*lconnection.ConnectorDetails.Version = getConnectorVersion(*lconnection.ConnectorVersion)
		} else {
			lconnection.ConnectorDetails.VersionId = new(string)
			*lconnection.ConnectorDetails.VersionId = getConnectorVersionId(*lconnection.ConnectorVersion)
		}

		lconnection.ConnectorVersion = nil
		fileName := getConnectionName(*lconnection.Name) + ".json"
		lconnection.Name = nil
		connectionPayload, err := json.Marshal(lconnection)
		if err != nil {
			return err
		}
		if err = apiclient.WriteByteArrayToFile(
			path.Join(apiclient.GetExportToFile(), fileName),
			false,
			connectionPayload); err != nil {
			clilog.Error.Println(err)
			return err
		}
		clilog.Info.Printf("Downloaded %s\n", fileName)
	}

	return nil
}

func RepairEvent(name string, wait bool) (err error) {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	u.Path = path.Join(u.Path, name)
	operationsBytes, err := apiclient.HttpClient(u.String(), "")
	if err != nil {
		return err
	}
	if wait {
		apiclient.ClientPrintHttpResponse.Set(false)
		defer apiclient.ClientPrintHttpResponse.Set(apiclient.GetCmdPrintHttpResponseSetting())

		o := operation{}
		if err = json.Unmarshal(operationsBytes, &o); err != nil {
			return err
		}

		operationId := filepath.Base(o.Name)
		clilog.Info.Printf("Checking connection repair status for %s in %d seconds\n", operationId, interval)

		stop := apiclient.Every(interval*time.Second, func(time.Time) bool {
			var respBody []byte

			if respBody, err = GetOperation(operationId); err != nil {
				return false
			}

			if err = json.Unmarshal(respBody, &o); err != nil {
				return false
			}

			if o.Done {
				if o.Error != nil {
					clilog.Error.Printf("Connection completed with error: %s\n", o.Error.Message)
				} else {
					clilog.Info.Println("Connection repair completed successfully!")
				}
				return false
			} else {
				clilog.Info.Printf("Connection repair status is: %t. Waiting %d seconds.\n", o.Done, interval)
				return true
			}
		})

		<-stop
	}
	return err
}

func getConnectorName(version string) string {
	return strings.Split(version, "/")[7]
}

func getConnectorVersion(version string) int {
	i, _ := strconv.Atoi(strings.Split(version, "/")[9])
	return i
}

func getConnectorVersionId(version string) string {
	return strings.Split(version, "/")[9]
}

func getConnectionName(name string) string {
	return name[strings.LastIndex(name, "/")+1:]
}

func getConnectorProvider(name string) string {
	return strings.Split(name, "/")[5]
}

func isGoogleConnection(connectionName string) bool {
	if connectionName == "pubsub" || connectionName == "gcs" || connectionName == "biqguery" ||
		connectionName == "cloudsql-mysql" || connectionName == "cloudsql-postgresql" ||
		connectionName == "cloudsql-sqlserver" || connectionName == "cloudspanner" {
		return true
	}
	return false
}
