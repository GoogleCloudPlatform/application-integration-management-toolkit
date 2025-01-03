// Copyright 2023 Google LLC
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

package endpoints

import (
	"fmt"
	"internal/apiclient"
	"internal/client/connections"
	"internal/clilog"
	"internal/cmd/utils"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CreateCmd to get endpoint attachments
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an endpoint attachments in the region",
	Long:  "Create an endpoint attachments in the region",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		project := utils.GetStringParam(cmd.Flag("proj"))
		region := utils.GetStringParam(cmd.Flag("reg"))

		if err = apiclient.SetRegion(region); err != nil {
			return err
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			clilog.Debug.Printf("%s: %s\n", f.Name, f.Value)
		})
		return apiclient.SetProjectID(project)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := utils.GetStringParam(cmd.Flag("name"))
		serviceAttachment := utils.GetStringParam(cmd.Flag("service-attachment"))
		description := utils.GetStringParam(cmd.Flag("description"))
		wait, _ := strconv.ParseBool(utils.GetStringParam(cmd.Flag("wait")))

		re := regexp.MustCompile(`projects\/([a-zA-Z0-9_-]+)\/regions` +
			`\/([a-zA-Z0-9_-]+)\/serviceAttachments\/([a-zA-Z0-9_-]+)`)

		if ok := re.Match([]byte(serviceAttachment)); !ok {
			return fmt.Errorf("The service attachment does not match the required format")
		}

		_, err = connections.CreateEndpoint(name, serviceAttachment, description, wait)
		return err
	},
}

func init() {
	var name, serviceAttachment, description string
	var wait bool

	CreateCmd.Flags().StringVarP(&name, "name", "n",
		"", "Endpoint attachment name; Ex: sample")
	CreateCmd.Flags().StringVarP(&serviceAttachment, "service-attachment", "s",
		"", "Endpoint attachment url; format = projects/*/regions/*/serviceAttachments/*")
	CreateCmd.Flags().StringVarP(&description, "description", "d",
		"", "Endpoint attachment description")
	CreateCmd.Flags().BoolVarP(&wait, "wait", "",
		false, "Waits for the connector to finish, with success or error; default is false")

	_ = CreateCmd.MarkFlagRequired("name")
	_ = CreateCmd.MarkFlagRequired("service-attachment")
}
