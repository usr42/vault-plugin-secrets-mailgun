package mgsecret

import (
	"context"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/helper/base62"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/hashicorp/vault/plugins/helper/database/credsutil"
	"github.com/mailgun/mailgun-go"
	"time"
)

const (
	SecretTypeSmtpCredentials = "smtp_credential_key"
	// TODO make user prefix configurable
	vaultUserPrefix           = "vault"
	internalDataUser          = "user_name"
)

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
	b.Logger().Info("Renewing credential", "username", req.Secret.InternalData[internalDataUser])
	resp := logical.Response{}
	resp.Secret = req.Secret
	resp.Secret.TTL = 1 * time.Minute
	resp.Secret.MaxTTL = 3 * time.Minute
	return &resp, nil
}

func (b *backend) secretCredentialsRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	username, ok := req.Secret.InternalData[internalDataUser]
	if !ok {
		return nil, fmt.Errorf("no internal user name found")
	}
	b.Logger().Info("Revoking credential", "username", username)
	// TODO: Refactor to reduce duplication
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, errwrap.Wrapf("Unable to get configuration: {{err}}", err)
	}

	if config == nil {
		return logical.ErrorResponse("Mailgun plugin is not configured."), nil
	}

	mgClient := mailgun.NewMailgun(config.Domain, config.ApiKey)

	if err = mgClient.DeleteCredential(username.(string)); err != nil {
		return logical.ErrorResponse(fmt.Sprintf("Unable to create credentials in mailgun: %v", err)), nil
	}

	return nil, nil
}

func (b *backend) generateCredentials(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, errwrap.Wrapf("Unable to get configuration: {{err}}", err)
	}

	if config == nil {
		return logical.ErrorResponse("Mailgun plugin is not configured."), nil
	}

	// TODO make password length configurable
	password, err := credsutil.RandomAlphaNumeric(32, false)
	if err != nil {
		return nil, errwrap.Wrapf("Unable to create random password: {{err}}", err)
	}

	userSuffix, err := base62.Random(5, true)
	if err != nil {
		return nil, errwrap.Wrapf("Unable to create unique, random username: {{err}}", err)
	}

	username := fmt.Sprintf("%v.%v", vaultUserPrefix, userSuffix)

	mgClient := mailgun.NewMailgun(config.Domain, config.ApiKey)

	if err = mgClient.CreateCredential(username, password); err != nil {
		return logical.ErrorResponse(fmt.Sprintf("Unable to create credentials in mailgun: %v", err)), nil
	}

	secretD := map[string]interface{}{
		"username": username,
		"password": password,
	}
	internalD := map[string]interface{}{
		internalDataUser: username,
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
