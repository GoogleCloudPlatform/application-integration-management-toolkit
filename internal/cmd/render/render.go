// Copyright 2025 Google LLC
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

package render

import (
	"internal/apiclient"
	"internal/cmd/utils"

	"github.com/spf13/cobra"
)

// Cmd to generate manifest and results
var Cmd = &cobra.Command{
	Use:   "render",
	Short: "Renders a default manifest.txt and results.json for Cloud Deploy",
	Long:  "Renders a default manifest.txt and results.json for Cloud Deploy",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		outputGCSPath := utils.GetStringParam(cmd.Flag("output-gcs-path"))
		v, _, _ := apiclient.GetBuildParams()
		err = apiclient.WriteManifest(outputGCSPath, v)
		return
	},
}

func init() {
	var outputGCSPath string

	Cmd.Flags().StringVarP(&outputGCSPath, "output-gcs-path", "",
		"", "Upload a file named results.json containing the results")

	_ = Cmd.MarkFlagRequired("output-gcs-path")
}
