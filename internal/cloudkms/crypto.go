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
	"bytes"
	"context"
	"encoding/base64"
	"internal/apiclient"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "cloud.google.com/go/kms/apiv1/kmspb"
)

// EncryptSymmetric will encrypt the input plaintext with the specified symmetric key.
func EncryptSymmetric(name string, plaintext []byte) (b64CipherText string, err error) {
	// kmsClient contains a client connection to cloud KMS
	var kmsClient *kms.KeyManagementClient

	ctx := context.Background()
	kmsClient, err = kms.NewKeyManagementClient(ctx)
	if err != nil {
		return "", apiclient.NewCliError("NewKeyManagementClient error", err)
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
		return "", apiclient.NewCliError("encrypt error", err)
	}

	// base64 encode the cipher
	b64CipherText = base64.StdEncoding.EncodeToString(resp.Ciphertext)

	return b64CipherText, nil
}

// DecryptSymmetric will decrypt the input ciphertext bytes using the specified symmetric key.
func DecryptSymmetric(name string, b64CipherText []byte) ([]byte, error) {
	// kmsClient contains a client connection to cloud KMS
	var kmsClient *kms.KeyManagementClient
	var err error

	ctx := context.Background()
	kmsClient, err = kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, apiclient.NewCliError("NewKeyManagementClient error", err)
	}

	defer kmsClient.Close()

	// base64 encode the cipher
	cipherText, err := base64.StdEncoding.DecodeString(string(b64CipherText))
	if err != nil {
		return nil, apiclient.NewCliError("decode base64 err", err)
	}

	// Build the request.
	req := &kmspb.DecryptRequest{
		Name:       name,
		Ciphertext: cipherText,
	}
	// Call the API.
	resp, err := kmsClient.Decrypt(ctx, req)
	if err != nil {
		return nil, apiclient.NewCliError("decrypt err", err)
	}

	return bytes.TrimSpace(resp.Plaintext), nil
}
