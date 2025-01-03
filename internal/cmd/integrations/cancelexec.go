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

package integrations

import (
	"internal/apiclient"
	"internal/client/integrations"
	"internal/clilog"
	"internal/cmd/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CancelExecCmd to list executions of an integration version
var CancelExecCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancels an execution of an integration",
	Long:  "Cancels an execution of an integration",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(utils.GetStringParam(cmdRegion)); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(utils.GetStringParam(cmdProject))
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		name := utils.GetStringParam(cmd.Flag("name"))
		executionID := utils.GetStringParam(cmd.Flag("execution-id"))
		cancelReason := utils.GetStringParam(cmd.Flag("cancel-reason"))
		_, err = integrations.Cancel(name, executionID, cancelReason)
		return err
	},
}

func init() {
	var name, executionID, cancelReason string

	CancelExecCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	CancelExecCmd.Flags().StringVarP(&cancelReason, "cancel-reason", "c",
		"", "Cancel Reason")
	CancelExecCmd.Flags().StringVarP(&executionID, "execution-id", "e",
		"", "Execution ID")

	_ = CancelExecCmd.MarkFlagRequired("name")
	_ = CancelExecCmd.MarkFlagRequired("cancel-reason")
	_ = CancelExecCmd.MarkFlagRequired("execution-id")
}
