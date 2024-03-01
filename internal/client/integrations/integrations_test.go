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
	"os"
	"path"
	"testing"

	"internal/client/clienttest"
	"internal/cmd/utils"
)

var cliPath = os.Getenv("INTEGRATIONCLI_PATH")

func TestCreateVersionNoOverrides(t *testing.T) {
	if err := clienttest.TestSetup(); err != nil {
		t.Fatalf("TestSetup failed: %v", err)
	}
	contents, err := utils.ReadFile(path.Join(cliPath, "test", "sample.json"))
	if err != nil {
		t.Fatalf("unable to read authConfig failed: %v", err)
	}
	_, err = CreateVersion("name", contents, nil, "", "")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestCreateVersionOverrides(t *testing.T) {
	if err := clienttest.TestSetup(); err != nil {
		t.Fatalf("TestSetup failed: %v", err)
	}
	contents, err := utils.ReadFile(path.Join(cliPath, "test", "sample.json"))
	if err != nil {
		t.Fatalf("unable to read authConfig failed: %v", err)
	}
	overrides, err := utils.ReadFile(path.Join(cliPath, "test", "sample_overrides.json"))
	if err != nil {
		t.Fatalf("unable to read authConfig failed: %v", err)
	}
	_, err = CreateVersion("name", contents, overrides, "2", "2")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestDelete(t *testing.T) {
	if err := clienttest.TestSetup(); err != nil {
		t.Fatalf("TestSetup failed: %v", err)
	}
	if _, err := Delete("test"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
