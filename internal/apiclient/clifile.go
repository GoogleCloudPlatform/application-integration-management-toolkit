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
	"os"
	"os/user"
	"path"
	"time"

	"internal/clilog"
)

const integrationcliFile = "config.json"
const integrationcliPath = ".integrationcli"

var usr *user.User

type integrationCLI struct {
	Token     string `json:"token,omitempty"`
	LastCheck string `json:"lastCheck,omitempty"`
	Project   string `json:"defaultProject,omitempty"`
	Region    string `json:"region,omitempty"`
	ProxyUrl  string `json:"proxyUrl,omitempty"`
	Nocheck   bool   `json:"nocheck,omitempty" default:"false"`
	UseApigee bool   `json:"useapigee,omitempty" default:"false"`
}

var cliPref *integrationCLI

func ReadPreferencesFile() (err error) {

	cliPref = new(integrationCLI)

	usr, err = user.Current()
	if err != nil {
		clilog.Info.Println(err)
		return err
	}

	prefFile, err := os.ReadFile(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile))
	if err != nil {
		clilog.Info.Println("Cached preferences was not found")
		return err
	}

	err = json.Unmarshal(prefFile, &cliPref)
	clilog.Info.Printf("Token %s, lastCheck: %s", cliPref.Token, cliPref.LastCheck)
	clilog.Info.Printf("DefaultProject %s", cliPref.Project)
	clilog.Info.Printf("Region %s", cliPref.Region)

	if err != nil {
		clilog.Info.Printf("Error marshalling: %v\n", err)
		return DeletePreferencesFile()
	}

	if cliPref.ProxyUrl != "" {
		SetProxyURL(cliPref.ProxyUrl)
	}

	if cliPref.Project != "" {
		if err = SetProjectID(cliPref.Project); err != nil {
			return err
		}
	}

	if cliPref.Region != "" {
		return SetRegion(cliPref.Region)
	}

	return nil
}

func DeletePreferencesFile() (err error) {
	usr, err = user.Current()
	if err != nil {
		clilog.Info.Println(err)
		return err
	}
	if _, err := os.Stat(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile)); os.IsNotExist(err) {
		clilog.Info.Println(err)
		return err
	}
	return os.Remove(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile))
}

func WriteToken(token string) (err error) {
	if IsSkipCache() {
		return nil
	}

	clilog.Info.Println("Cache access token: ", token)
	if cliPref == nil {
		clilog.Info.Printf("preferences are not set")
		return nil
	}

	cliPref.Token = token

	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Info.Printf("Error marshalling: %v\n", err)
		return err
	}
	clilog.Info.Println("Writing ", string(data))
	return WritePerferencesFile(data)
}

func GetToken() (token string) {
	if cliPref == nil {
		return ""
	}
	return cliPref.Token
}

func GetLastCheck() (lastCheck string) {
	if cliPref == nil {
		return ""
	}
	return cliPref.LastCheck
}

func GetNoCheck() bool {
	if cliPref == nil {
		return false
	}
	return cliPref.Nocheck
}

func SetNoCheck(nocheck bool) (err error) {
	clilog.Info.Println("Nocheck set to: ", nocheck)

	if cliPref == nil {
		cliPref = &integrationCLI{}
	}

	cliPref.Nocheck = nocheck

	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Info.Printf("Error marshalling: %v\n", err)
		return err
	}
	clilog.Info.Println("Writing ", string(data))
	return WritePerferencesFile(data)
}

func SetUseApigee(useapigee bool) (err error) {
	clilog.Info.Println("UseApigee set to: ", useapigee)

	if cliPref == nil {
		cliPref = &integrationCLI{}
	}

	cliPref.UseApigee = false
	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Info.Printf("Error marshalling: %v\n", err)
		return err
	}
	clilog.Info.Println("Writing ", string(data))
	return WritePerferencesFile(data)
}

func TestAndUpdateLastCheck() (updated bool, err error) {
	currentTime := time.Now()
	currentDate := currentTime.Format("01-02-2006")

	if cliPref == nil {
		cliPref = &integrationCLI{}
	}

	if currentDate == cliPref.LastCheck {
		return true, nil
	}

	cliPref.LastCheck = currentDate

	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Warning.Printf("Error marshalling: %v\n", err)
		return false, err
	}
	clilog.Info.Println("Writing ", string(data))
	if err = WritePerferencesFile(data); err != nil {
		return false, err
	}

	return false, nil
}

func GetDefaultProject() (org string) {
	return cliPref.Project
}

func WriteDefaultProject(project string) (err error) {
	clilog.Info.Println("Default project: ", project)
	if cliPref == nil {
		cliPref = &integrationCLI{}
	}
	cliPref.Project = project
	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Info.Printf("Error marshalling: %v\n", err)
		return err
	}
	clilog.Info.Println("Writing ", string(data))
	return WritePerferencesFile(data)
}

func SetProxy(url string) (err error) {
	if url == "" {
		return nil
	}
	if cliPref == nil {
		cliPref = &integrationCLI{}
	}
	cliPref.ProxyUrl = url
	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Info.Printf("Error marshalling: %v\n", err)
		return err
	}
	clilog.Info.Println("Writing ", string(data))
	return WritePerferencesFile(data)
}

func SetDefaultRegion(region string) (err error) {
	if region == "" {
		return nil
	}
	if cliPref == nil {
		cliPref = &integrationCLI{}
	}
	cliPref.Region = region
	data, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Info.Printf("Error marshalling: %v\n", err)
		return err
	}
	clilog.Info.Println("Writing ", string(data))
	return WritePerferencesFile(data)
}

func GetPreferences() (err error) {
	output, err := json.Marshal(&cliPref)
	if err != nil {
		clilog.Error.Println(err)
		return err
	}

	PrettyPrint(output)
	return nil
}

// WritePreferencesFile
func WritePerferencesFile(payload []byte) (err error) {
	usr, err = user.Current()
	if err != nil {
		clilog.Warning.Println(err)
		return err
	}
	_, err = os.Stat(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile))
	if err == nil {
		return WriteByteArrayToFile(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile), false, payload)
	} else if os.IsNotExist(err) {
		if err = os.MkdirAll(path.Join(usr.HomeDir, integrationcliPath), 0755); err != nil {
			return err
		}
		return WriteByteArrayToFile(path.Join(usr.HomeDir, integrationcliPath, integrationcliFile), false, payload)
	} else if err != nil {
		clilog.Warning.Println(err)
		return err
	}
	return nil
}
