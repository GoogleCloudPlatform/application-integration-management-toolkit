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

package main

import (
	"internal/cmd/authconfigs"
	"internal/cmd/connectors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	cmd "internal/cmd"

	integrations "internal/cmd/integrations"

	"github.com/spf13/cobra/doc"

	apiclient "internal/apiclient"
)

const ENABLED = "true"

var samples = `# integrationcli command Samples

The following table contains some examples of integrationcli.

Set up integrationcli with preferences: ` + getSingleLine("integrationcli prefs set -p $project -r $region") + `

| Operations | Command |
|---|---|
| integrations | ` + getSingleLine(integrations.GetExample(0)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(1)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(2)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(13)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(3)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(4)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(5)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(6)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(7)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(8)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(9)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(10)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(11)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(12)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(14)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(15)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(16)) + `|
| integrations | ` + getSingleLine(integrations.GetExample(17)) + `|
| authconfigs | ` + getSingleLine(authconfigs.GetExample(0)) + `|
| authconfigs | ` + getSingleLine(authconfigs.GetExample(1)) + `|
| authconfigs | ` + getSingleLine(authconfigs.GetExample(2)) + `|
| authconfigs | ` + getSingleLine(authconfigs.GetExample(3)) + `|
| connectors | ` + getSingleLine(connectors.GetExample(0)) + `|


NOTE: This file is auto-generated during a release. Do not modify.`

func main() {
	var err error
	var docFiles []string

	if os.Getenv("INTEGRATIONCLI_SKIP_DOCS") != ENABLED {

		if docFiles, err = filepath.Glob("./docs/integrationcli*.md"); err != nil {
			log.Fatal(err)
		}

		for _, docFile := range docFiles {
			if err = os.Remove(docFile); err != nil {
				log.Fatal(err)
			}
		}

		if err = doc.GenMarkdownTree(cmd.RootCmd, "./docs"); err != nil {
			log.Fatal(err)
		}
	}

	_ = apiclient.WriteByteArrayToFile("./samples/README.md", false, []byte(samples))
}

func WriteFile() (byteValue []byte, err error) {
	userFile, err := os.Open("./samples/README.md")
	if err != nil {
		return nil, err
	}

	defer userFile.Close()

	byteValue, err = io.ReadAll(userFile)
	if err != nil {
		return nil, err
	}
	return byteValue, err
}

func getSingleLine(s string) string {
	return "`" + strings.ReplaceAll(strings.ReplaceAll(s, "\\", ""), "\n", "") + "`"
}
