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
	"fmt"
	"internal/apiclient"
	"internal/cmd"
	"os"
)

// https://goreleaser.com/cookbooks/using-main.version/?h=ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := cmd.GetRootCmd()
	apiclient.SetBuildParams(version, commit, date)
	rootCmd.Version = fmt.Sprintf("%s date: %s [commit: %.7s]", version, date, commit)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
