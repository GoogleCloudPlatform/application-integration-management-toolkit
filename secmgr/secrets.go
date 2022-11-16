package secmgr

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

func Create(project string, secretId string, payload []byte) (err error) {
	ctx := context.Background()

	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	// Create the request to create the secret.
	req := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", project),
		SecretId: secretId,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	secret, err := c.CreateSecret(ctx, req)
	if err != nil {
		return err
	}

	// Build the request.
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: payload,
		},
	}

	// Call the API.
	_, err = c.AddSecretVersion(ctx, addSecretVersionReq)
	if err != nil {
		return err
	}

	return nil
}
