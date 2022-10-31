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

package cloudkms

import (
	"context"
	"encoding/base64"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

// EncryptSymmetric will encrypt the input plaintext with the specified symmetric key.
func EncryptSymmetric(name string, plaintext []byte) (b64CipherText string, err error) {

	// kmsClient contains a client connection to cloud KMS
	var kmsClient *kms.KeyManagementClient

	ctx := context.Background()
	kmsClient, err = kms.NewKeyManagementClient(ctx)
	if err != nil {
		return "", err
	}

	defer kmsClient.Close()

	// Build the request.
	req := &kmspb.EncryptRequest{
		Name:      name,
		Plaintext: plaintext,
	}

	// Call the API.
	resp, err := kmsClient.Encrypt(ctx, req)
	if err != nil {
		return "", fmt.Errorf("encrypt error: %v", err)
	}

	//base64 encode the cipher
	b64CipherText = base64.StdEncoding.EncodeToString(resp.Ciphertext)

	return b64CipherText, nil
}
