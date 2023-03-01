// Copyright 2021 Google LLC
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

// VerCmd to manage versions of an integration flow
var VerCmd = &cobra.Command{
	Use:   "versions",
	Short: "Manage integrations flow versions",
	Long:  "Manage integrations flow versions",
}

var version string

func init() {

	VerCmd.AddCommand(ListVerCmd)
	VerCmd.AddCommand(PatchVerCmd)
	VerCmd.AddCommand(GetVerCmd)
	VerCmd.AddCommand(ExportVerCmd)
	VerCmd.AddCommand(ImportflowCmd)
	VerCmd.AddCommand(PublishVerCmd)
	VerCmd.AddCommand(UnPublishVerCmd)
	VerCmd.AddCommand(DownloadVerCmd)
	VerCmd.AddCommand(DelVerCmd)
}
