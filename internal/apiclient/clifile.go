// Copyright 2020 Google LLC
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
	"internal/clilog"
	"os"
	"os/user"
	"path"
	"strconv"
	"time"
)

const (
	integrationcliFile = "config.json"
	integrationcliPath = ".integrationcli"
)

type integrationCLI struct {
	Token     string `json:"token,omitempty"`
	LastCheck string `json:"lastCheck,omitempty"`
	Project   string `json:"defaultProject,omitempty"`
	Region    string `json:"region,omitempty"`
	ProxyUrl  string `json:"proxyUrl,omitempty"`
	Nocheck   bool   `json:"nocheck,omitempty" default:"false"`
	Api       API    `json:"api,omitempty" default:"prod"`
	BasicInfo string `json:"basicInfo,omitempty" default:"false"`
}

func readPreferencesFile() (cliPref *integrationCLI, err error) {
	cliPref = new(integrationCLI)

	usr, err := user.Current()
	if err != nil {
		clilog.Debug.Println(err)
		return cliPref, newError("unable to get current user", err)
	}

	prefFile, err := os.ReadFile(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile))
	if err != nil {
		clilog.Debug.Println("Cached preferences was not found")
		return cliPref, newError("Cached preferences was not found", err)
	}

	err = json.Unmarshal(prefFile, &cliPref)
	clilog.Debug.Printf("Token %s, lastCheck: %s", cliPref.Token, cliPref.LastCheck)
	clilog.Debug.Printf("DefaultProject %s", cliPref.Project)
	clilog.Debug.Printf("Region %s", cliPref.Region)

	if err != nil {
		clilog.Debug.Printf("Error marshalling: %v\n", err)
		return cliPref, DeletePreferencesFile()
	}

	if cliPref.Api != "" {
		if cliPref.Api != PROD && cliPref.Api != STAGING && cliPref.Api != AUTOPUSH {
			return cliPref, newError("invalid api settings in configuration file", nil)
		}
	}

	return cliPref, nil
}

func DeletePreferencesFile() (err error) {
	usr, err := user.Current()
	if err != nil {
		clilog.Debug.Println(err)
		return newError("unable to get current user", err)
	}
	if _, err := os.Stat(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile)); os.IsNotExist(err) {
		clilog.Debug.Println(err)
		return newError("file does not exist", err)
	}
	return os.Remove(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile))
}

func writeToken(token string) (err error) {
	if IsSkipCache() {
		return nil
	}

	cliPref, err := readPreferencesFile()
	if err != nil {
		return err
	}

	clilog.Debug.Println("Cache access token: ", token)
	cliPref.Token = token

	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Debug.Printf("Error marshalling: %v\n", err)
		return newError("error marshalling", err)
	}
	clilog.Debug.Println("Writing ", string(data))
	return writePerferencesFile(data)
}

func getToken() (token string) {
	cliPref, err := readPreferencesFile()
	if err != nil {
		return ""
	}
	return cliPref.Token
}

func GetNoCheck() bool {
	cliPref, err := readPreferencesFile()
	if err != nil {
		return false
	}
	return cliPref.Nocheck
}

func SetNoCheck(nocheck bool) (err error) {
	clilog.Debug.Println("Nocheck set to: ", nocheck)

	cliPref, err := readPreferencesFile()
	cliPref.Nocheck = nocheck

	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Debug.Printf("Error marshalling: %v\n", err)
		return newError("error marshalling", err)
	}
	clilog.Debug.Println("Writing ", string(data))
	return writePerferencesFile(data)
}

func GetBasicInfo() string {
	cliPref, err := readPreferencesFile()
	if err != nil {
		return ""
	}
	return cliPref.BasicInfo
}

func SetBasicInfo(basicInfo string) (err error) {
	clilog.Debug.Println("BasicInfo set to: ", basicInfo)

	if _, err = strconv.ParseBool(basicInfo); err != nil {
		return newError("error parsing to bool", err)
	}

	clilog.Debug.Println("BasicInfo set to: ", basicInfo)

	cliPref, err := readPreferencesFile()
	cliPref.BasicInfo = basicInfo

	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Debug.Printf("Error marshalling: %v\n", err)
		return newError("error marshalling", err)
	}
	clilog.Debug.Println("Writing ", string(data))
	return writePerferencesFile(data)
}

func SetAPIPref(a API) (err error) {
	clilog.Debug.Println("API is set to: ", a)

	cliPref, err := readPreferencesFile()

	cliPref.Api = a
	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Debug.Printf("Error marshalling: %v\n", err)
		return newError("error marshalling", err)
	}
	clilog.Debug.Println("Writing ", string(data))
	return writePerferencesFile(data)
}

func TestAndUpdateLastCheck() (updated bool, err error) {
	currentTime := time.Now()
	currentDate := currentTime.Format("01-02-2006")

	cliPref, err := readPreferencesFile()
	clilog.Debug.Println(err)

	if currentDate == cliPref.LastCheck {
		return true, nil
	}

	cliPref.LastCheck = currentDate

	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Warning.Printf("Error marshalling: %v\n", err)
		return false, newError("error marshalling", err)
	}
	clilog.Debug.Println("Writing ", string(data))
	if err = writePerferencesFile(data); err != nil {
		return false, newError("error writing preference", err)
	}
	return false, nil
}

func GetDefaultProject() string {
	cliPref, _ := readPreferencesFile()
	return cliPref.Project
}

func WriteDefaultProject(project string) (err error) {
	clilog.Debug.Println("Default project: ", project)
	cliPref, err := readPreferencesFile()
	cliPref.Project = project
	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Debug.Printf("Error marshalling: %v\n", err)
		return newError("error marshalling", err)
	}
	clilog.Debug.Println("Writing ", string(data))
	return writePerferencesFile(data)
}

func SetProxy(url string) (err error) {
	if url == "" {
		return nil
	}
	cliPref, err := readPreferencesFile()
	cliPref.ProxyUrl = url
	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Debug.Printf("Error marshalling: %v\n", err)
		return newError("error marshalling", err)
	}
	clilog.Debug.Println("Writing ", string(data))
	return writePerferencesFile(data)
}

func SetDefaultRegion(region string) (err error) {
	if region == "" {
		return nil
	}
	cliPref, err := readPreferencesFile()
	cliPref.Region = region
	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Debug.Printf("Error marshalling: %v\n", err)
		return newError("error marshalling", err)
	}
	clilog.Debug.Println("Writing ", string(data))
	return writePerferencesFile(data)
}

func GetPreferences() APIResponse {
	cliPref, err := readPreferencesFile()
	if err != nil {
		clilog.Error.Println(err)
		return APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	output, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Error.Println(err)
		return APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	return APIResponse{
		RespBody: output,
		Err:      nil,
	}
}

// WritePreferencesFile
func writePerferencesFile(payload []byte) (err error) {
	usr, err := user.Current()
	if err != nil {
		clilog.Warning.Println(err)
		return err
	}
	_, err = os.Stat(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile))
	if err == nil {
		return WriteByteArrayToFile(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile), false, payload)
	} else {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(path.Join(usr.HomeDir, integrationcliPath), 0o755); err != nil {
				return err
			}
			return WriteByteArrayToFile(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile), false, payload)
		} else {
			clilog.Warning.Println(err)
			return err
		}
	}
}
