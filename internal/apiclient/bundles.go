// Copyright 2020 Google LLC
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

package apiclient

import (
	"bytes"
	"fmt"
	"os"

	"internal/clilog"
)

// entityPayloadList stores list of entities
var entityPayloadList [][]byte //types.EntityPayloadList

// WriteByteArrayToFile accepts []bytes and writes to a file
func WriteByteArrayToFile(exportFile string, fileAppend bool, payload []byte) error {
	var fileFlags = os.O_CREATE | os.O_WRONLY

	if fileAppend {
		fileFlags |= os.O_APPEND
	} else {
		fileFlags |= os.O_TRUNC
	}

	f, err := os.OpenFile(exportFile, fileFlags, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(payload)
	if err != nil {
		clilog.Error.Println("error writing to file: ", err)
		return err
	}
	return nil
}

// WriteArrayByteArrayToFile accepts [][]bytes and writes to a file
func WriteArrayByteArrayToFile(exportFile string, fileAppend bool, payload [][]byte) error {
	var fileFlags = os.O_CREATE | os.O_WRONLY

	if fileAppend {
		fileFlags |= os.O_APPEND
	}

	f, err := os.OpenFile(exportFile, fileFlags, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	//begin json array
	_, err = f.Write([]byte("["))
	if err != nil {
		clilog.Error.Println("error writing to file ", err)
		return err
	}

	payloadFromArray := bytes.Join(payload, []byte(","))
	//add json array terminate
	payloadFromArray = append(payloadFromArray, byte(']'))

	_, err = f.Write(payloadFromArray)

	if err != nil {
		clilog.Error.Println("error writing to file: ", err)
		return err
	}

	return nil
}

func FolderExists(folder string) (err error) {
	if folder == "" {
		return nil
	}
	_, err = os.Stat(folder)
	if err != nil {
		return fmt.Errorf("folder not found or write permission denied")
	}
	return nil
}

func SetEntityPayloadList(respBody []byte) {
	entityPayloadList = append(entityPayloadList, respBody)
}

func GetEntityPayloadList() [][]byte {
	return entityPayloadList
}

func ClearEntityPayloadList() {
	entityPayloadList = entityPayloadList[:0]
}
