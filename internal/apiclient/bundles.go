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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"internal/clilog"

	"cloud.google.com/go/storage"
)

// entityPayloadList stores list of entities
var entityPayloadList [][]byte // types.EntityPayloadList

// WriteByteArrayToFile accepts []bytes and writes to a file
func WriteByteArrayToFile(exportFile string, fileAppend bool, payload []byte) error {
	fileFlags := os.O_CREATE | os.O_WRONLY

	if fileAppend {
		fileFlags |= os.O_APPEND
	} else {
		fileFlags |= os.O_TRUNC
	}

	f, err := os.OpenFile(exportFile, fileFlags, 0o644)
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
	fileFlags := os.O_CREATE | os.O_WRONLY

	if fileAppend {
		fileFlags |= os.O_APPEND
	}

	f, err := os.OpenFile(exportFile, fileFlags, 0o644)
	if err != nil {
		return err
	}

	defer f.Close()

	// begin json array
	_, err = f.Write([]byte("["))
	if err != nil {
		clilog.Error.Println("error writing to file ", err)
		return err
	}

	payloadFromArray := bytes.Join(payload, []byte(","))
	// add json array terminate
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

func ExtractTgz(gcsURL string) (folder string, err error) {

	ctx := context.Background()

	folder, err = os.MkdirTemp("", "integration")
	fmt.Println(folder)
	if err != nil {
		return "", err
	}

	// Parse the GCS URL
	parsedURL, err := url.Parse(gcsURL)
	if err != nil {
		return "", fmt.Errorf("Error parsing GCS URL:", err)
	}
	if parsedURL.Scheme != "gs" {
		return "", fmt.Errorf("Invalid GCS URL scheme. Should be 'gs://'")
	}

	bucketName := parsedURL.Host
	objectName := strings.TrimPrefix(parsedURL.Path, "/")
	fileName := filepath.Base(gcsURL)

	// Create a Google Cloud Storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("Error creating GCS client:", err)
	}
	defer client.Close()

	// Get a handle to the bucket and the object
	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)

	// Create a reader to stream the object's content
	reader, err := object.NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("Error creating object reader:", err)
	}
	defer reader.Close()

	// Create the local file to save the download
	localFile, err := os.Create(path.Join(folder, fileName))
	if err != nil {
		return "", fmt.Errorf("Error creating local file:", err)
	}
	defer localFile.Close()

	// Download the object and save it to the local file
	if _, err := io.Copy(localFile, reader); err != nil {
		return "", fmt.Errorf("Error downloading object:", err)
	}

	// Open the .tgz file
	file, err := os.Open(path.Join(folder, fileName))
	if err != nil {
		return "", fmt.Errorf("Error opening file:", err)
	}
	defer file.Close() // Ensure file closure

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("Error creating gzip reader:", err)
	}
	defer gzipReader.Close() // Ensure closure

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract each file from the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return "", fmt.Errorf("Error reading tar entry:", err)
		}

		// Process the file header
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.Mkdir(path.Join(folder, header.Name), 0755); err != nil {
				return "", fmt.Errorf("Error creating directory:", err)
			}
		case tar.TypeReg:
			// Create output file
			outFile, err := os.Create(path.Join(folder, header.Name))
			if err != nil {
				return "", fmt.Errorf("Error creating file:", err)
			}
			defer outFile.Close()

			// Copy contents from the tar to the output file
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return "", fmt.Errorf("Error writing file:", err)
			}
		default:
			return "", fmt.Errorf("Unsupported type: %b in %s\n", header.Typeflag, header.Name)
		}
	}
	return folder, nil
}

func GetSkaffoldConfigUri(pipeline string, release string) (uri string, err error) {
	cloudDeployRespMap := make(map[string]interface{})
	cloudDeployURL := fmt.Sprintf("https://clouddeploy.googleapis.com/v1/projects/%s/locations/%s/deliveryPipelines/%s/releases/%s",
		GetProjectID(), GetRegion(), pipeline, release)
	u, _ := url.Parse(cloudDeployURL)

	ClientPrintHttpResponse.Set(false)

	respBody, err := HttpClient(u.String())
	if err != nil {
		return "", err
	}
	defer ClientPrintHttpResponse.Set(GetCmdPrintHttpResponseSetting())

	err = json.Unmarshal(respBody, &cloudDeployRespMap)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", cloudDeployRespMap["skaffoldConfigUri"]), nil
}
