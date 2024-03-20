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
		if strings.Contains(header.Name, "..") {
			continue
		}

		// Process the file header
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.Mkdir(path.Join(folder, header.Name), 0o755); err != nil {
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

func GetCloudDeployGCSLocations(pipeline string, release string) (skaffoldConfigUri string, err error) {
	type cloudDeployRelease struct {
		SkaffoldConfigUri string `json:"skaffoldConfigUri"`
		TargetArtifacts   map[string]struct {
			SkaffoldConfigPath string `json:"skaffoldConfigPath"`
			ManifestPath       string `json:"manifestPath"`
			ArtifactUri        string `json:"artifactUri"`
			PhaseArtifacts     map[string]struct {
				SkaffoldConfigPath string `json:"skaffoldConfigPath"`
				ManifestPath       string `json:"manifestPath"`
			} `json:"phaseArtifacts"`
		} `json:"targetArtifacts"`
	}

	r := cloudDeployRelease{}

	cloudDeployURL := fmt.Sprintf("https://clouddeploy.googleapis.com/v1/projects/%s/locations/%s/deliveryPipelines/%s/releases/%s",
		GetProjectID(), GetRegion(), pipeline, release)
	u, _ := url.Parse(cloudDeployURL)

	ClientPrintHttpResponse.Set(false)

	respBody, err := HttpClient(u.String())
	if err != nil {
		return "", err
	}
	defer ClientPrintHttpResponse.Set(GetCmdPrintHttpResponseSetting())

	err = json.Unmarshal(respBody, &r)
	if err != nil {
		return "", err
	}

	return r.SkaffoldConfigUri, nil
}

func WriteResultsFile(deployOutputGCS string, status string) (err error) {
	contents := fmt.Sprintf("{\"resultStatus\": \"%s\"}", status)
	filename := "results.json"

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Extract bucket name and object path from GCS URI
	bucketName, objectPath, err := parseGCSURI(deployOutputGCS)
	objectName := path.Join(objectPath, filename)

	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)
	writer := object.NewWriter(ctx)

	// Write the content
	if _, err := writer.Write([]byte(contents)); err != nil {
		return fmt.Errorf("Object(%q).NewWriter: %v", objectName, err)
	}

	// Close the writer to ensure data is uploaded
	if err := writer.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

func parseGCSURI(gcsURI string) (bucketName, objectPath string, err error) {
	// Parse the GCS URL
	parsedURL, err := url.Parse(gcsURI)
	if err != nil {
		return "", "", fmt.Errorf("Error parsing GCS URL:", err)
	}
	if parsedURL.Scheme != "gs" {
		return "", "", fmt.Errorf("Invalid GCS URL scheme. Should be 'gs://'")
	}
	// Remove the protocol prefix
	uri := strings.TrimPrefix(gcsURI, "gs://")

	// Split based on the first '/'
	parts := strings.SplitN(uri, "/", 2)

	// Check for proper URI format
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid GCS URI format")
	}
	return parts[0], parts[1], nil
}
