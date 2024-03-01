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

package authconfigs

import (
	"os"
	"path"
	"testing"

	"internal/client/clienttest"
	"internal/cmd/utils"
)

var cliPath = os.Getenv("INTEGRATIONCLI_PATH")

var authConfigID string

func TestCreate(t *testing.T) {
	var contents []byte
	if err := clienttest.TestSetup(); err != nil {
		t.Fatalf("TestSetup failed: %v", err)
	}
	contents, err := utils.ReadFile(path.Join(cliPath, "test", "ac_username.json"))
	if err != nil {
		t.Fatalf("unable to read authConfig failed: %v", err)
	}
	if _, err := Create(contents); err != nil {
		t.Fatalf("TestCreate failed: %v", err)
	}
}

func TestFind(t *testing.T) {
	var err error
	if err = clienttest.TestSetup(); err != nil {
		t.Fatalf("TestSetup failed: %v", err)
	}
	authConfigID, err = Find("authconfig-sample", "")
	if err != nil {
		t.Fatalf("TestFind failed: %v", err)
	}
}

func TestGet(t *testing.T) {
	if err := clienttest.TestSetup(); err != nil {
		t.Fatalf("TestSetup failed: %v", err)
	}
	if _, err := Get(authConfigID, false); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
}

func TestGetDisplayName(t *testing.T) {
	if err := clienttest.TestSetup(); err != nil {
		t.Fatalf("TestSetup failed: %v", err)
	}
	if _, err := GetDisplayName("authconfig-sample"); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
}

func TestList(t *testing.T) {
	if err := clienttest.TestSetup(); err != nil {
		t.Fatalf("TestSetup failed: %v", err)
	}
	if _, err := List(-1, "", ""); err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestDelete(t *testing.T) {
	if err := clienttest.TestSetup(); err != nil {
		t.Fatalf("TestSetup failed: %v", err)
	}
	if _, err := Delete(authConfigID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
