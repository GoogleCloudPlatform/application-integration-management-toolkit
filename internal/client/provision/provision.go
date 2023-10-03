// Copyright 2023 Google LLC
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

package provision

import (
	"fmt"
	"internal/apiclient"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// Provision
func Provision(cloudkms string, samples bool, gmek bool, serviceAccount string) (respBody []byte, err error) {
	u, _ := url.Parse(apiclient.GetBaseIntegrationURL())
	provStr := []string{}

	if serviceAccount != "" {
		provStr = append(provStr, "\"runAsServiceAccount\":\""+serviceAccount+"\"")
	}

	if cloudkms != "" {
		kmsConfig := getCloudKMSConfig(cloudkms)
		provStr = append(provStr, "\"cloudKmsConfig\":"+kmsConfig)
	}

	provStr = append(provStr, "\"createSampleWorkflows\":"+strconv.FormatBool(samples))
	provStr = append(provStr, "\"provisionGmek\":"+strconv.FormatBool(gmek))

	u.Path = path.Join(u.Path, "client:provision")

	payload := "{" + strings.Join(provStr, ",") + "}"
	fmt.Println(payload)
	respBody, err = apiclient.HttpClient(u.String(), payload)
	return respBody, err
}

func getCloudKMSConfig(cloudkms string) string {
	kmsParts := strings.Split(cloudkms, "/")
	return fmt.Sprintf("{\"kmsLocation\":\"%s\",\"kmsRing\":\"%s\",\"key\":\"%s\",\"keyVersion\":\"%s\",\"kmsProjectId\":\"%s\"}",
		kmsParts[3], kmsParts[5], kmsParts[7], kmsParts[9], kmsParts[1])
}
