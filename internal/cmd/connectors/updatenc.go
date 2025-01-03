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

package connectors

import (
	"errors"
	"fmt"
	"internal/apiclient"
	"internal/client/connections"
	"internal/clilog"
	"internal/cmd/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// UpdateNodeCountCmd to get connection
var UpdateNodeCountCmd = &cobra.Command{
	Use:   "update",
	Short: "Update connection max or min node count",
	Long:  "Update connection max or min node count",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		cmdProject := cmd.Flag("proj")
		cmdRegion := cmd.Flag("reg")

		if err = apiclient.SetRegion(utils.GetStringParam(cmdRegion)); err != nil {
			return err
		}
		if min == -1 && max == -1 {
			return errors.New("min or max must be set")
		}
		if min == 0 || max == 0 {
			return errors.New("min or max cannot be set to 0")
		}
		if min > max && max != -1 {
			return errors.New("min cannot be set higher than max")
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(utils.GetStringParam(cmdProject))
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		nodeConfig := []string{}
		var nodeCount string
		name := utils.GetStringParam(cmd.Flag("name"))

		content := "{\"nodeConfig\": {"

		if min != -1 && max != -1 {
			nodeConfig = append(nodeConfig, "nodeConfig.minNodeCount")
			nodeConfig = append(nodeConfig, "nodeConfig.maxNodeCount")
			nodeCount = fmt.Sprintf("\"maxNodeCount\": %d", max) + "," + fmt.Sprintf("\"minNodeCount\": %d", min)
		} else if min != -1 {
			nodeConfig = append(nodeConfig, "nodeConfig.minNodeCount")
			nodeCount = fmt.Sprintf("\"minNodeCount\": %d", min)
		} else if max != -1 {
			nodeConfig = append(nodeConfig, "nodeConfig.maxNodeCount")
			nodeCount = fmt.Sprintf("\"maxNodeCount\": %d", max)
		}

		content = content + nodeCount + "}}"
		_, err = connections.Patch(name, []byte(content), nodeConfig)
		return err
	},
}

var min, max int

func init() {
	var name string

	UpdateNodeCountCmd.Flags().StringVarP(&name, "name", "n",
		"", "The name of the connection")
	UpdateNodeCountCmd.Flags().IntVarP(&min, "min", "",
		-1, "Min node count for a connection")
	UpdateNodeCountCmd.Flags().IntVarP(&max, "max", "",
		-1, "Max node count for a connection")

	_ = UpdateNodeCountCmd.MarkFlagRequired("name")
}
