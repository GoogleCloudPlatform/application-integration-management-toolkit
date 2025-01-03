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

// ReplayExecCmd to list executions of an integration version
var ReplayExecCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replays an execution of an integration",
	Long:  "Replays an execution of an integration",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(cmdRegion.Value.String()); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(cmdProject.Value.String())
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := utils.GetStringParam(cmd.Flag("name"))
		executionID := utils.GetStringParam(cmd.Flag("execution-id"))
		replayReason := utils.GetStringParam(cmd.Flag("replay-reason"))
		_, err = integrations.Replay(name, executionID, replayReason)
		return err
	},
}

func init() {
	var name, executionID, replayReason string

	ReplayExecCmd.Flags().StringVarP(&name, "name", "n",
		"", "Integration flow name")
	ReplayExecCmd.Flags().StringVarP(&replayReason, "replay-reason", "",
		"", "Replay Reason")
	ReplayExecCmd.Flags().StringVarP(&executionID, "execution-id", "e",
		"", "Execution ID")

	_ = ReplayExecCmd.MarkFlagRequired("name")
	_ = ReplayExecCmd.MarkFlagRequired("replay-reason")
	_ = ReplayExecCmd.MarkFlagRequired("execution-id")
}
