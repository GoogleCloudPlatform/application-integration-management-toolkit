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

package apiclient

import (
	"encoding/json"
	"fmt"
	"internal/clilog"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

// condition for Bindings
type condition struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Expression  string `json:"expression,omitempty"`
}

// binding for IAM Roles
type roleBinding struct {
	Role      string     `json:"role,omitempty"`
	Members   []string   `json:"members,omitempty"`
	Condition *condition `json:"condition,omitempty"`
}

// IamPolicy holds the response
type iamPolicy struct {
	Version  int           `json:"version,omitempty"`
	Etag     string        `json:"etag,omitempty"`
	Bindings []roleBinding `json:"bindings,omitempty"`
}

// SetIamPolicy holds the request to set IAM
type setIamPolicy struct {
	Policy iamPolicy `json:"policy,omitempty"`
}

func iamServiceAccountExists(iamname string) (code int, err error) {
	var resp *http.Response
	var req *http.Request

	projectid, _, err := getNameAndProject(iamname)
	if err != nil {
		return -1, NewCliError("unable to get project id", err)
	}

	getendpoint := fmt.Sprintf("https://iam.googleapis.com/v1/projects/%s/serviceAccounts/%s", projectid, iamname)
	contentType := "application/json"

	client, err := getHttpClient()
	if err != nil {
		return -1, NewCliError("unable to get http client", err)
	}

	if DryRun() {
		return 200, nil
	}

	req, err = http.NewRequest("GET", getendpoint, nil)
	if err != nil {
		return -1, NewCliError("error in client", err)
	}

	req, err = setAuthHeader(req)
	if err != nil {
		return -1, NewCliError("error setting auth header", err)
	}

	clilog.Debug.Println("Content-Type : ", contentType)
	req.Header.Set("Content-Type", contentType)

	resp, err = client.Do(req)
	if err != nil {
		return resp.StatusCode, NewCliError("error connecting", err)
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if resp == nil {
		return -1, NewCliError("error in response: Response was null", nil)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, NewCliError("error in response", err)
	} else if resp.StatusCode > 399 && resp.StatusCode != 404 {
		return resp.StatusCode, NewCliError("error in client", fmt.Errorf("status code %d, error in response: %s", resp.StatusCode, string(respBody)))
	} else {
		return resp.StatusCode, nil
	}
}

// setIAMPermission set permissions for a member
func setIAMPermission(endpoint string, name string, memberName string, role string, memberType string) (err error) {
	u, _ := url.Parse(endpoint)
	u.Path = path.Join(u.Path, name+":getIamPolicy")

	response := HttpClient(u.String())
	if response.Err != nil {
		return NewCliError("error in http client", response.Err)
	}

	getIamPolicy := iamPolicy{}

	err = json.Unmarshal(response.RespBody, &getIamPolicy)
	if err != nil {
		return NewCliError("error marshalling", err)
	}

	foundRole := false
	for i, binding := range getIamPolicy.Bindings {
		if binding.Role == role {
			// found members with the role already, add the new SA to the role
			getIamPolicy.Bindings[i].Members = append(binding.Members, memberType+":"+memberName)
			foundRole = true
		}
	}

	// no members with the role, add a new one
	if !foundRole {
		binding := roleBinding{}
		binding.Role = role
		binding.Members = append(binding.Members, memberType+":"+memberName)
		getIamPolicy.Bindings = append(getIamPolicy.Bindings, binding)
	}

	u, _ = url.Parse(endpoint)
	u.Path = path.Join(u.Path, name+":setIamPolicy")

	setIamPolicy := setIamPolicy{}
	setIamPolicy.Policy = getIamPolicy

	setIamPolicyBody, err := json.Marshal(setIamPolicy)
	if err != nil {
		return NewCliError("error marshalling", err)
	}

	response = HttpClient(u.String(), string(setIamPolicyBody))
	if response.Err != nil {
		return NewCliError("error in http client", response.Err)
	}
	return nil
}

// setProjectIAMPermission
func setProjectIAMPermission(project string, memberName string, role string) (err error) {
	getendpoint := fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v1/projects/%s:getIamPolicy", project)
	setendpoint := fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v1/projects/%s:setIamPolicy", project)

	// this method treats errors as info since this is not a blocking problem

	// Get the current IAM policies for the project
	response := HttpClient(getendpoint, "")
	if response.Err != nil {
		return NewCliError("error in http client", response.Err)
	}

	// binding for IAM Roles
	type roleBinding struct {
		Role      string     `json:"role,omitempty"`
		Members   []string   `json:"members,omitempty"`
		Condition *condition `json:"condition,omitempty"`
	}

	// IamPolicy holds the response
	type iamPolicy struct {
		Version  int           `json:"version,omitempty"`
		Etag     string        `json:"etag,omitempty"`
		Bindings []roleBinding `json:"bindings,omitempty"`
	}

	// iamPolicyRequest holds the request to set IAM
	type iamPolicyRequest struct {
		Policy iamPolicy `json:"policy,omitempty"`
	}

	policy := iamPolicy{}

	err = json.Unmarshal(response.RespBody, &policy)
	if err != nil {
		return NewCliError("error unmarshalling", err)
	}

	binding := roleBinding{}
	binding.Role = role
	binding.Members = append(binding.Members, "serviceAccount:"+memberName)

	policy.Bindings = append(policy.Bindings, binding)

	policyRequest := iamPolicyRequest{}
	policyRequest.Policy = policy
	policyRequestBody, err := json.Marshal(policyRequest)
	if err != nil {
		return NewCliError("error marshalling", err)
	}

	response = HttpClient(setendpoint, string(policyRequestBody))
	if response.Err != nil {
		clilog.Debug.Printf("error setting IAM policies for the project %s: %v", project, err)
		return NewCliError("error in http client", response.Err)
	}

	return nil
}

// CreateServiceAccount
func CreateServiceAccount(iamname string) (err error) {
	var statusCode int

	projectid, displayname, err := getNameAndProject(iamname)
	if err != nil {
		return NewCliError("unable to get project id", err)
	}

	if statusCode, err = iamServiceAccountExists(iamname); err != nil {
		return NewCliError("unable to fetch service account details", err)
	}

	switch statusCode {
	case 200:
		return nil
	case 404:
		createendpoint := fmt.Sprintf("https://iam.googleapis.com/v1/projects/%s/serviceAccounts", projectid)
		iamPayload := []string{}
		iamPayload = append(iamPayload, "\"accountId\":\""+displayname+"\"")
		iamPayload = append(iamPayload, "\"serviceAccount\": {\"displayName\": \""+displayname+"\"}")
		payload := "{" + strings.Join(iamPayload, ",") + "}"

		if response := HttpClient(createendpoint, payload); response.Err != nil {
			return NewCliError("error in http client", response.Err)
		}
		return nil
	default:
		return NewCliError("unhandled status code", fmt.Errorf("unable to fetch service account details, err: %d", statusCode))
	}
}

// SetConnectorIAMPermission set permissions for a member on a connection
func SetConnectorIAMPermission(name string, memberName string, iamRole string, memberType string) (err error) {
	var role string

	switch iamRole {
	case "admin":
		role = "roles/connectors.admin"
	case "invoker":
		role = "roles/connectors.invoker"
	case "viewer":
		role = "roles/connectors.viewer"
	default: // assume this is a custom role definition
		re := regexp.MustCompile(`projects\/([a-zA-Z0-9_-]+)\/roles\/([a-zA-Z0-9_-]+)`)
		result := re.FindString(iamRole)
		if result == "" {
			return fmt.Errorf("custom role must be of the format projects/{project-id}/roles/{role-name}")
		}
		role = iamRole
	}

	return setIAMPermission(GetBaseConnectorURL(), name, memberName, role, memberType)
}

// SetPubSubIAMPermission set permissions for a SA on a topic
func SetPubSubIAMPermission(project string, topic string, memberName string) (err error) {
	endpoint := fmt.Sprintf("https://pubsub.googleapis.com/v1/projects/%s/topics", project)
	const memberType = "serviceAccount"
	const role = "roles/pubsub.publisher"
	return setIAMPermission(endpoint, topic, memberName, role, memberType)
}

// SetSecretManagerIAMPermission set permissions for a SA on a secret
func SetSecretManagerIAMPermission(project string, secretName string, memberName string) (err error) {
	endpoint := fmt.Sprintf("https://secretmanager.googleapis.com/v1/projects/%s/secrets", project)
	const memberType = "serviceAccount"
	const role1 = "roles/secretmanager.secretAccessor"
	const role2 = "roles/secretmanager.viewer"
	if err = setIAMPermission(endpoint, secretName, memberName, role1, memberType); err != nil {
		return err
	}
	return setIAMPermission(endpoint, secretName, memberName, role2, memberType)
}

// SetBigQueryIAMPermission
func SetBigQueryIAMPermission(project string, datasetid string, memberName string) (err error) {
	endpoint := fmt.Sprintf("https://bigquery.googleapis.com/bigquery/v2/projects/%s/datasets/%s", project, datasetid)
	const role = "WRITER"
	var content []byte

	// first fetch the information
	response := HttpClient(endpoint)
	if response.Err != nil {
		return NewCliError("error in http client", response.Err)
	}

	type accessType struct {
		Role         string  `json:"role,omitempty"`
		IamMember    *string `json:"iamMember,omitempty"`
		UserByEmail  *string `json:"userByEmail,omitempty"`
		SpecialGroup *string `json:"specialGroup,omitempty"`
		GroupByEmail *string `json:"groupByEmail,omitempty"`
	}

	type datasetType struct {
		Access []accessType `json:"access,omitempty"`
	}

	dataset := datasetType{}
	if err = json.Unmarshal(response.RespBody, &dataset); err != nil {
		return NewCliError("error unmarshalling", err)
	}

	access := accessType{}
	access.Role = role
	access.UserByEmail = new(string)
	*access.UserByEmail = memberName

	// merge the updates
	dataset.Access = append(dataset.Access, access)

	if content, err = json.Marshal(dataset); err != nil {
		return NewCliError("error marshalling", err)
	}

	// patch the update
	if response := HttpClient(endpoint, string(content), "PATCH"); response.Err != nil {
		return NewCliError("error in http client", response.Err)
	}

	return nil
}

// SetCloudStorageIAMPermission
func SetCloudStorageIAMPermission(project string, memberName string) (err error) {
	// the connector currently requires storage.buckets.list. other built-in roles didn't have this permission
	const role = "roles/storage.admin"

	return setProjectIAMPermission(project, memberName, role)
}

// SetCloudSQLIAMPermission
func SetCloudSQLIAMPermission(project string, memberName string) (err error) {
	const role = "roles/cloudsql.editor"
	return setProjectIAMPermission(project, memberName, role)
}

// SetCloudSpannerIAMPermission
func SetCloudSpannerIAMPermission(project string, memberName string) (err error) {
	const role = "roles/spanner.databaseUser"
	return setProjectIAMPermission(project, memberName, role)
}

// SetIntegrationInvokerPermission
func SetIntegrationInvokerPermission(project string, memberName string) (err error) {
	const role = "roles/integrations.integrationInvoker"
	return setProjectIAMPermission(project, memberName, role)
}

func getNameAndProject(iamFullName string) (projectid string, name string, err error) {
	riam := regexp.MustCompile(`^[a-zA-Z0-9-]{6,30}$`)

	parts := strings.Split(iamFullName, "@")

	if len(parts) != 2 {
		return "", "", NewCliError("error splitting iam name", fmt.Errorf("invalid iam name, %s", parts))
	}
	name = parts[0]
	projectid = strings.ReplaceAll(parts[1], ".iam.gserviceaccount.com", "") // strings.Split(parts[1], ".iam.gserviceaccount.com")[0]
	if name == "" || projectid == "" {
		return "", "", NewCliError("error splitting iam name", fmt.Errorf("invalid iam name %s, %s", name, projectid))
	}
	if ok := riam.Match([]byte(name)); !ok {
		return "", "", NewCliError("error splitting iam name", fmt.Errorf("the ID must be between 6 and 30 characters"))
	}
	return projectid, name, nil
}

// GetDefaultServiceAccount
func GetComputeEngineDefaultServiceAccount(projectId string) (serviceAccount string, err error) {
	getendpoint := fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v3/projects/%s", projectId)

	// Get the project number

	response := HttpClient(getendpoint)
	if response.Err != nil {
		clilog.Debug.Printf("error getting details for the project %s: %v", projectId, err)
		return serviceAccount, NewCliError("error in http client", response.Err)
	}

	type projectResponse struct {
		Name        string            `json:"name,omitempty"`
		Parent      string            `json:"parent,omitempty"`
		ProjectId   string            `json:"projectId,omitempty"`
		State       string            `json:"state,omitempty"`
		DisplayName string            `json:"displayName,omitempty"`
		CreateTime  string            `json:"createTime,omitempty"`
		UpdateTime  string            `json:"updateTime,omitempty"`
		DeleteTime  string            `json:"deleteTime,omitempty"`
		Etag        string            `json:"etag,omitempty"`
		Labels      map[string]string `json:"labels,omitempty"`
	}

	p := projectResponse{}

	err = json.Unmarshal(response.RespBody, &p)
	if err != nil {
		clilog.Debug.Println(err)
		return serviceAccount, NewCliError("error unmarshalling", err)
	}

	if p.Name == "" {
		return serviceAccount, NewCliError("error getting project numner", fmt.Errorf("project number was not available"))
	}

	// get the project number
	projectNumber := strings.Split(p.Name, "/")[1]
	serviceAccount = fmt.Sprintf("%s-compute@developer.gserviceaccount.com", projectNumber)

	return serviceAccount, nil
}
