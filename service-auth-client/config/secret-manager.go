package config

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type SecretConfig struct {
	Context    *context.Context
	ProjectID  string
	SecretName string
	Version    string
}

func InitSecretConfig(ctx *context.Context, projectId, secretName, version string) *SecretConfig {
	return &SecretConfig{
		Context:    ctx,
		ProjectID:  projectId,
		SecretName: secretName,
		Version:    version,
	}
}

func (sc *SecretConfig) AccessSecretResource(client *secretmanager.Client) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", sc.ProjectID, sc.SecretName, sc.Version),
	}

	secret, err := client.AccessSecretVersion(*sc.Context, accessRequest)
	if err != nil {
		return nil, err
	}

	return secret, nil
}
