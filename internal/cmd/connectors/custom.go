// Copyright 2024 Google LLC
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

package connectors

import (
	"github.com/spf13/cobra"
)

// CustomCmd to manage preferences
var CustomCmd = &cobra.Command{
	Use:   "custom",
	Short: "Manage custom connections for Integration Connectors",
	Long:  "Manage custom connections for Integration Connectors",
}

func init() {
	CustomCmd.AddCommand(GetCustomCmd)
	CustomCmd.AddCommand(ListCustomCmd)
	CustomCmd.AddCommand(DelCustomCmd)
	CustomCmd.AddCommand(CrtCustomCmd)
	CustomCmd.AddCommand(CustomVerCmd)
}
