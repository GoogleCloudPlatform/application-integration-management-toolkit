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
	"fmt"
	"net/url"
	"path"

	"github.com/srinandan/integrationcli/apiclient"
)

var validMemberTypes = []string{"serviceAccount", "group", "user", "domain"}

// GetIAM
func GetIAM(name string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	u.Path = path.Join(u.Path, name+":getIamPolicy")
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String())
	return respBody, err
}

// SetIAM
func SetIAM(name string, memberName string, permission string, memberType string) (err error) {
	if !isValidMemberType(memberType) {
		return fmt.Errorf("invalid memberType. Valid types are %v", validMemberTypes)
	}
	return apiclient.SetIAMPermission(name, memberName, permission, memberType)
}

// TestIAM
func TestIAM(name string, resource string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseConnectorURL())
	u.Path = path.Join(u.Path, name+":testIamPermissions")
	payload := "{\"permissions\":[\"" + resource + "\"]}"
	respBody, err = apiclient.HttpClient(apiclient.GetPrintOutput(), u.String(), payload)
	return respBody, err
}

func isValidMemberType(memberType string) bool {
	for _, validMember := range validMemberTypes {
		if memberType == validMember {
			return true
		}
	}
	return false
}
