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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/apigee/apigeecli/clilog"
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
		clilog.Error.Println(err)
		return -1, err
	}

	var getendpoint = fmt.Sprintf("https://iam.googleapis.com/v1/projects/%s/serviceAccounts/%s", projectid, iamname)
	var contentType = "application/json"

	client, err := getHttpClient()
	if err != nil {
		clilog.Error.Println(err)
		return -1, err
	}

	if DryRun() {
		return 200, nil
	}

	req, err = http.NewRequest("GET", getendpoint, nil)
	if err != nil {
		clilog.Error.Println("error in client: ", err)
		return -1, err
	}

	req, err = setAuthHeader(req)
	if err != nil {
		clilog.Error.Println(err)
		return -1, err
	}

	clilog.Info.Println("Content-Type : ", contentType)
	req.Header.Set("Content-Type", contentType)

	resp, err = client.Do(req)
	if err != nil {
		clilog.Error.Println("error connecting: ", err)
		return resp.StatusCode, err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if resp == nil {
		return -1, fmt.Errorf("error in response: Response was null")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		clilog.Error.Println("error in response: ", err)
		return -1, err
	} else if resp.StatusCode > 399 && resp.StatusCode != 404 {
		clilog.Error.Printf("status code %d, error in response: %s\n", resp.StatusCode, string(respBody))
		return resp.StatusCode, errors.New("error in response")
	} else {
		return resp.StatusCode, nil
	}
}

// setIAMPermission set permissions for a member
func setIAMPermission(endpoint string, name string, memberName string, role string, memberType string) (err error) {

	u, _ := url.Parse(endpoint)
	u.Path = path.Join(u.Path, name+":getIamPolicy")
	getIamPolicyBody, err := HttpClient(false, u.String())
	if err != nil {
		clilog.Error.Println(err)
		return err
	}

	getIamPolicy := iamPolicy{}

	err = json.Unmarshal(getIamPolicyBody, &getIamPolicy)
	if err != nil {
		clilog.Error.Println(err)
		return err
	}

	foundRole := false
	for i, binding := range getIamPolicy.Bindings {
		if binding.Role == role {
			//found members with the role already, add the new SA to the role
			getIamPolicy.Bindings[i].Members = append(binding.Members, memberType+":"+memberName)
			foundRole = true
		}
	}

	//no members with the role, add a new one
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
		clilog.Error.Println(err)
		return err
	}

	_, err = HttpClient(false, u.String(), string(setIamPolicyBody))

	return err
}

// setProjectIAMPermission
func setProjectIAMPermission(project string, memberName string, role string) (err error) {
	var getendpoint = fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v1/projects/%s:getIamPolicy", project)
	var setendpoint = fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v1/projects/%s:setIamPolicy", project)

	//this method treats errors as info since this is not a blocking problem

	//Get the current IAM policies for the project
	respBody, err := HttpClient(false, getendpoint, "")
	if err != nil {
		clilog.Info.Printf("error getting IAM policies for the project %s: %v", project, err)
		return err
	}

	//binding for IAM Roles
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

	//iamPolicyRequest holds the request to set IAM
	type iamPolicyRequest struct {
		Policy iamPolicy `json:"policy,omitempty"`
	}

	policy := iamPolicy{}

	err = json.Unmarshal(respBody, &policy)
	if err != nil {
		clilog.Info.Println(err)
		return err
	}

	binding := roleBinding{}
	binding.Role = role
	binding.Members = append(binding.Members, "serviceAccount:"+memberName)

	policy.Bindings = append(policy.Bindings, binding)

	policyRequest := iamPolicyRequest{}
	policyRequest.Policy = policy
	policyRequestBody, err := json.Marshal(policyRequest)
	if err != nil {
		clilog.Info.Println(err)
		return err
	}

	_, err = HttpClient(false, setendpoint, string(policyRequestBody))
	if err != nil {
		clilog.Info.Printf("error setting IAM policies for the project %s: %v", project, err)
		return err
	}

	return nil
}

// CreateServiceAccount
func CreateServiceAccount(iamname string) (err error) {

	var statusCode int

	projectid, displayname, err := getNameAndProject(iamname)
	if err != nil {
		return err
	}

	if statusCode, err = iamServiceAccountExists(iamname); err != nil {
		return err
	}

	switch statusCode {
	case 200:
		return nil
	case 404:
		var createendpoint = fmt.Sprintf("https://iam.googleapis.com/v1/projects/%s/serviceAccounts", projectid)
		iamPayload := []string{}
		iamPayload = append(iamPayload, "\"accountId\":\""+displayname+"\"")
		iamPayload = append(iamPayload, "\"serviceAccount\": {\"displayName\": \""+displayname+"\"}")
		payload := "{" + strings.Join(iamPayload, ",") + "}"

		if _, err = HttpClient(false, createendpoint, payload); err != nil {
			clilog.Error.Println(err)
			return err
		}
		return nil
	default:
		return fmt.Errorf("unable to fetch service account details, err: %d", statusCode)
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
	default: //assume this is a custom role definition
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
	var endpoint = fmt.Sprintf("https://pubsub.googleapis.com/v1/projects/%s/topics", project)
	const memberType = "serviceAccount"
	const role = "roles/pubsub.publisher"
	return setIAMPermission(endpoint, topic, memberName, role, memberType)
}

// SetSecretManagerIAMPermission set permissions for a SA on a secret
func SetSecretManagerIAMPermission(project string, secretName string, memberName string) (err error) {
	var endpoint = fmt.Sprintf("https://secretmanager.googleapis.com/v1/projects/%s/secrets", project)
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
	var endpoint = fmt.Sprintf("https://bigquery.googleapis.com/bigquery/v2/projects/%s/datasets/%s", project, datasetid)
	const role = "WRITER"
	var content []byte

	//first fetch the information
	respBody, err := HttpClient(false, endpoint)
	if err != nil {
		return err
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
	if err = json.Unmarshal(respBody, &dataset); err != nil {
		return err
	}

	access := accessType{}
	access.Role = role
	access.UserByEmail = new(string)
	*access.UserByEmail = memberName

	//merge the updates
	dataset.Access = append(dataset.Access, access)

	if content, err = json.Marshal(dataset); err != nil {
		return err
	}

	//patch the update
	if _, err = HttpClient(false, endpoint, string(content), "PATCH"); err != nil {
		return err
	}

	return nil
}

// SetCloudStorageIAMPermission
func SetCloudStorageIAMPermission(project string, memberName string) (err error) {
	//the connector currently requires storage.buckets.list. other built-in roles didn't have this permission
	const role = "roles/storage.admin"

	return setProjectIAMPermission(project, memberName, role)
}

// SetCloudSQLIAMPermission
func SetCloudSQLIAMPermission(project string, memberName string) (err error) {
	const role = "roles/cloudsql.editor"
	return setProjectIAMPermission(project, memberName, role)
}

func getNameAndProject(iamFullName string) (projectid string, name string, err error) {

	riam := regexp.MustCompile(`^[a-zA-Z0-9-]{6,30}$`)

	parts := strings.Split(iamFullName, "@")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid iam name")
	}
	name = parts[0]
	projectid = strings.Split(parts[1], ".iam.gserviceaccount.com")[0]
	if name == "" || projectid == "" {
		return "", "", fmt.Errorf("invalid iam name")
	}
	if ok := riam.Match([]byte(name)); !ok {
		return "", "", fmt.Errorf("the ID must be between 6 and 30 characters")
	}
	return projectid, name, nil
}

// GetDefaultServiceAccount
func GetComputeEngineDefaultServiceAccount(projectId string) (serviceAccount string, err error) {
	var getendpoint = fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v3/projects/%s", projectId)

	//Get the project number
	respBody, err := HttpClient(false, getendpoint, "")
	if err != nil {
		clilog.Info.Printf("error getting details for the project %s: %v", projectId, err)
		return serviceAccount, err
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

	err = json.Unmarshal(respBody, &p)
	if err != nil {
		clilog.Info.Println(err)
		return serviceAccount, err
	}

	if p.Name == "" {
		return serviceAccount, fmt.Errorf("project number was not available")
	}

	//get the project number
	projectNumber := strings.Split(p.Name, "/")[1]
	serviceAccount = fmt.Sprintf("%s-compute@developer.gserviceaccount.com", projectNumber)

	return serviceAccount, nil
}
