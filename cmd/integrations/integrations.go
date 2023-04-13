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

package integrations

import (
	"github.com/spf13/cobra"
)

// Cmd to manage integration flows
var Cmd = &cobra.Command{
	Use:   "integrations",
	Short: "Manage integrations in a GCP project",
	Long:  "Manage integrations in a GCP project",
}

var userLabel, snapshot string
var overrides bool

func init() {

	var project, region string

	Cmd.PersistentFlags().StringVarP(&project, "proj", "p",
		"", "Integration GCP Project name")

	Cmd.PersistentFlags().StringVarP(&region, "reg", "r",
		"", "Integration region name")

	Cmd.AddCommand(ListCmd)
	Cmd.AddCommand(VerCmd)
	Cmd.AddCommand(CleanCmd)
	Cmd.AddCommand(ExecuteCmd)
	Cmd.AddCommand(ExecCmd)
	Cmd.AddCommand(ExportCmd)
	Cmd.AddCommand(ImportCmd)
	Cmd.AddCommand(UploadCmd)
	Cmd.AddCommand(CreateCmd)
	Cmd.AddCommand(DelCmd)
	Cmd.AddCommand(ScaffoldCmd)
	Cmd.AddCommand(ApplyCmd)
}
