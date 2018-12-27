package mgsecret

import (
	"context"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/helper/base62"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/hashicorp/vault/plugins/helper/database/credsutil"
	"strings"
)

const (
	SecretTypeSmtpCredentials = "smtp_credential_key"
	vaultUserPrefix           = "vault"
	internalDataUser          = "user_name"
)

func pathCredentials(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "credentials",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.generateCredentials,
				Summary:  "Get mailgun SMTP username and password.",
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
				Description: "The SMTP username for mailgun.",
			},
			"password": {
				Type:        framework.TypeString,
				Description: "The SMTP password for mailgun.",
			},
		},
		Renew:  b.secretCredentialsRenew,
		Revoke: b.secretCredentialsRevoke,
	}
}

func (b *backend) secretCredentialsRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Info("Renewing credential", "username", req.Secret.InternalData[internalDataUser])
	config, err := getConfig(ctx, req.Storage)
	if ok, response, err := handleGetConfig(err, config); !ok {
		return response, err
	}
	resp := logical.Response{}
	resp.Secret = req.Secret
	resp.Secret.TTL = config.TTL
	resp.Secret.MaxTTL = config.MaxTTL
	return &resp, nil
}

func (b *backend) secretCredentialsRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	username, ok := req.Secret.InternalData[internalDataUser]
	if !ok {
		return nil, fmt.Errorf("no internal user name found")
	}

	b.Logger().Info("Revoking credential", "username", username)

	config, err := getConfig(ctx, req.Storage)
	if ok, response, err := handleGetConfig(err, config); !ok {
		return response, err
	}

	client := b.MailgunFactory(config.Domain, config.ApiKey)

	if err = client.DeleteCredential(username.(string)); err != nil {
		return logical.ErrorResponse("Unable to create credentials in mailgun. Configure with valid credentials"), nil
	}

	return nil, nil
}

func (b *backend) generateCredentials(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if ok, response, err := handleGetConfig(err, config); !ok {
		return response, err
	}

	password, err := credsutil.RandomAlphaNumeric(32, false)
	if err != nil {
		return nil, errwrap.Wrapf("Unable to create random password: {{err}}", err)
	}

	userSuffix, err := base62.Random(5, true)
	if err != nil {
		return nil, errwrap.Wrapf("Unable to create unique, random username: {{err}}", err)
	}
	userSuffix = strings.ToLower(userSuffix)

	username := fmt.Sprintf("%v.%v", vaultUserPrefix, userSuffix)

	client := b.MailgunFactory(config.Domain, config.ApiKey)

	if err = client.CreateCredential(username, password); err != nil {
		return logical.ErrorResponse(fmt.Sprintf("Unable to create credentials in mailgun: %v", err)), nil
	}

	secretD := map[string]interface{}{
		"username": fmt.Sprintf("%s@%s", username, config.Domain),
		"password": password,
	}
	internalD := map[string]interface{}{
		internalDataUser: username,
	}

	secret := b.Secret(SecretTypeSmtpCredentials)
	resp := secret.Response(secretD, internalD)
	resp.Secret.TTL = config.TTL
	resp.Secret.MaxTTL = config.MaxTTL
	resp.Secret.Renewable = true
	return resp, nil
}

func handleGetConfig(err error, config *config) (bool, *logical.Response, error) {
	if err != nil {
		return false, nil, errwrap.Wrapf("Unable to get configuration: {{err}}", err)
	}
	if config == nil {
		return false, logical.ErrorResponse("Mailgun plugin is not configured."), nil
	}
	return true, nil, nil
}
