package secmgr

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/srinandan/integrationcli/apiclient"
)

// secretExists the latest secret version
func secretExists(project string, name string) (version string, err error) {

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.GetSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", project, name),
	}

	// Call the API.
	secretVersion, err := client.GetSecretVersion(ctx, req)
	if err != nil {
		return "", err
	}

	return secretVersion.Name, nil
}

// Create a new secret in secret manager
func Create(project string, secretId string, payload []byte) (version string, err error) {

	if version, err = secretExists(project, secretId); err == nil {
		return version, nil //secret exists, return
	}

	ctx := context.Background()

	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer c.Close()

	//secret manager location
	replica := &secretmanagerpb.Replication_UserManaged_Replica{}
	replica.Location = apiclient.GetRegion()

	replicas := []*secretmanagerpb.Replication_UserManaged_Replica{}
	replicas = append(replicas, replica)

	// Create the request to create the secret.
	req := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", project),
		SecretId: secretId,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_UserManaged_{
					UserManaged: &secretmanagerpb.Replication_UserManaged{
						Replicas: replicas,
					},
				},
			},
		},
	}

	secret, err := c.CreateSecret(ctx, req)
	if err != nil {
		return "", err
	}

	// Build the request.
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: payload,
		},
	}

	// Call the API.
	secretVersion, err := c.AddSecretVersion(ctx, addSecretVersionReq)
	if err != nil {
		return "", err
	}

	return secretVersion.Name, nil
}
