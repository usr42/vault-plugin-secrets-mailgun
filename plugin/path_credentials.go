package mgsecret

import (
	"context"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/hashicorp/vault/plugins/helper/database/credsutil"
	"time"
)

const SecretTypeSmtpCredentials = "smtp_credential_key"

func pathCredentials(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "credentials",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.generateCredentials,
				Summary:  "TODO.",
			},
		},
	}
}

func secretCredentials(b *backend) *framework.Secret {
	return &framework.Secret{
		Type: SecretTypeSmtpCredentials,
		Fields: map[string]*framework.FieldSchema{
			"username": {
				Type:        framework.TypeString,
				Description: "TODO",
			},
			"password": {
				Type:        framework.TypeString,
				Description: "TODO",
			},
		},
		Renew:  b.secretCredentialsRenew,
		Revoke: b.secretCredentialsRevoke,
	}
}

func (b *backend) secretCredentialsRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Info("Renewing Secret", "key_name", req.Secret.InternalData["key_name"])
	resp := logical.Response{}
	resp.Secret = req.Secret
	resp.Secret.TTL = 1 * time.Minute
	resp.Secret.MaxTTL = 3 * time.Minute
	return &resp, nil
}

func (b *backend) secretCredentialsRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Info("Revoking Secret", "key_name", req.Secret.InternalData["key_name"])
	return nil, nil
}

func (b *backend) generateCredentials(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	// TODO Use real credentials
	randomAlphaNumeric, err := credsutil.RandomAlphaNumeric(10, false)
	if err != nil {
		return nil, err
	}
	secretD := map[string]interface{}{
		"username": randomAlphaNumeric,
		"password": randomAlphaNumeric,
	}
	internalD := map[string]interface{}{
		"key_name": randomAlphaNumeric,
	}

	secret := b.Secret(SecretTypeSmtpCredentials)
	resp := secret.Response(secretD, internalD)
	// TODO Use TTL and MaxTTL from config
	// TODO Use same TTL values for renew
	resp.Secret.TTL = 1 * time.Minute
	resp.Secret.MaxTTL = 3 * time.Minute
	resp.Secret.Renewable = true
	return resp, nil
}
