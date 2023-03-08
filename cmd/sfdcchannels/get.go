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

package sfdcchannels

import (
	"github.com/srinandan/integrationcli/apiclient"
	"github.com/srinandan/integrationcli/client/sfdc"

	"github.com/spf13/cobra"
)

// GetCmd to get integration flow
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an sfdcchannel in Application Integration",
	Long:  "Get an sfdcchannel in Application Integration",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if name != "" {
			apiclient.SetPrintOutput(false)
			_, respBody, err := sfdc.FindChannel(name, instance)
			if err != nil {
				return err
			}
			apiclient.PrettyPrint(respBody)
		} else {
			_, err = sfdc.GetChannel(id, instance, minimal)
		}
		return

	},
}

var minimal bool
var id string

func init() {
	GetCmd.Flags().StringVarP(&name, "name", "n",
		"", "sfdc channel name")
	GetCmd.Flags().StringVarP(&id, "id", "i",
		"", "sfdc channel uuid")
	GetCmd.Flags().StringVarP(&instance, "instance", "",
		"", "sfdc instance uuid")
	GetCmd.Flags().BoolVarP(&minimal, "minimal", "",
		false, "Minimal number of fields returned; default is false")

	_ = GetCmd.MarkFlagRequired("instance")
}
