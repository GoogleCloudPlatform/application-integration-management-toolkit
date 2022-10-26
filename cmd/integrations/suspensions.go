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

// SuspendCmd to manage suspensions of an integration flow
var SuspendCmd = &cobra.Command{
	Use:     "suspensions",
	Aliases: []string{"susp"},
	Short:   "Manage suspensions of an integrations flow version",
	Long:    "Manage suspensions of an integrations flow version",
}

func init() {
	SuspendCmd.AddCommand(ListSuspCmd)
	SuspendCmd.AddCommand(LiftSuspCmd)
	//SuspendCmd.AddCommand(ResolveSuspendCmd)
}
