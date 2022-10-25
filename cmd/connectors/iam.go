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

package connectors

import (
	"github.com/spf13/cobra"
)

// IamCmd to IAM permissions for Connections
var IamCmd = &cobra.Command{
	Use:   "iam",
	Short: "Manage IAM permissions for the connection",
	Long:  "Manage IAM permissions for the connection",
}

var memberName, role, memberType string

func init() {

	IamCmd.PersistentFlags().StringVarP(&name, "name", "n",
		"", "Connection name")

	_ = IamCmd.MarkPersistentFlagRequired("name")

	IamCmd.AddCommand(GetIamCmd)
	IamCmd.AddCommand(SetAdminCmd)
	//IamCmd.AddCommand(SetInvokeCmd)
	//IamCmd.AddCommand(SetViewerCmd)
	//IamCmd.AddCommand(SetCustCmd)
}
