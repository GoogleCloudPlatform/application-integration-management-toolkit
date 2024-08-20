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

package sfdcinstances

import (
	"errors"
	"internal/apiclient"
	"internal/client/sfdc"
	"strconv"

	"github.com/spf13/cobra"
)

// GetCmd to get integration flow
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an sfdcinstance in Application Integration",
	Long:  "Get an sfdcinstance in Application Integration",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		project := cmd.Flag("proj").Value.String()
		region := cmd.Flag("reg").Value.String()
		name := cmd.Flag("name").Value.String()
		id := cmd.Flag("id").Value.String()

		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		if id == "" && name == "" {
			return errors.New("id and name cannot be empty")
		}
		if id != "" && name != "" {
			return errors.New("id and name both cannot be set")
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := cmd.Flag("name").Value.String()
		id := cmd.Flag("id").Value.String()
		minimal, _ := strconv.ParseBool(cmd.Flag("minimal").Value.String())

		if name != "" {
			apiclient.DisableCmdPrintHttpResponse()
			_, respBody, err := sfdc.FindInstance(name)
			if err != nil {
				return err
			}
			apiclient.EnableCmdPrintHttpResponse()
			apiclient.PrettyPrint(respBody)
		} else {
			_, err = sfdc.GetInstance(id, minimal)
		}
		return err
	},
}

func init() {
	var name, id string
	minimal := false

	GetCmd.Flags().StringVarP(&id, "id", "i",
		"", "Instance name (uuid)")
	GetCmd.Flags().StringVarP(&name, "name", "n",
		"", "Instance display name")
	GetCmd.Flags().BoolVarP(&minimal, "minimal", "",
		false, "Minimal number of fields returned; default is false")

	_ = GetCmd.MarkFlagRequired("name")
}
