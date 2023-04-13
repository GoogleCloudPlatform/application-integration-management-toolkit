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
	"errors"
	"internal/apiclient"

	"internal/client/integrations"

	"github.com/spf13/cobra"
)

// LiftSuspCmd to list suspensions of an integration
var LiftSuspCmd = &cobra.Command{
	Use:   "lift",
	Short: "Lift a suspension of an integration execution",
	Long:  "Lift a suspension of an integration execution",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return errors.Unwrap(err)
		}
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := cmd.Flag("name").Value.String()
		_, err = integrations.Lift(name, execution, suspension, result)
		return

	},
}

var execution, suspension, result string

func init() {
	var name string

	LiftSuspCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration name")
	LiftSuspCmd.Flags().StringVarP(&execution, "execution", "e",
		"", "Integration execution id")
	LiftSuspCmd.Flags().StringVarP(&suspension, "suspension", "s",
		"", "Integration suspension id")
	LiftSuspCmd.Flags().StringVarP(&result, "result", "",
		"", "Integration suspension result")

	_ = LiftSuspCmd.MarkFlagRequired("name")
	_ = LiftSuspCmd.MarkFlagRequired("execution")
	_ = LiftSuspCmd.MarkFlagRequired("suspension")
	_ = LiftSuspCmd.MarkFlagRequired("result")
}
