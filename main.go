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
	"os"
	"runtime/debug"

	"github.com/GoogleCloudPlatform/application-integration-management-toolkit/cmd"
)

// Version
var Version string

func main() {

	rootCmd := cmd.GetRootCmd()
	versionDetails := ""

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if setting.Value != "" {
					versionDetails = versionDetails + ", git: " + setting.Value
				}
			case "vcs.time":
				if setting.Value != "" {
					versionDetails = versionDetails + ", time: " + setting.Value
				}
			}
		}
	}

	if Version == "" {
		Version = "(not set)"
	}

	rootCmd.Version = Version + versionDetails

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
