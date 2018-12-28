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
		HelpSynopsis:    pathCredentialsSyn,
		HelpDescription: pathCredentialsDesc,
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
	resp := logical.Response{Secret: req.Secret}
	return &resp, nil
}

func (b *backend) secretCredentialsRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	username, ok := req.Secret.InternalData[internalDataUser]
	if !ok {
		return nil, fmt.Errorf("no internal user name found")
	}

	config, err := getConfig(ctx, req.Storage)
	if ok, response, err := handleGetConfigErrors(err, config); !ok {
		return response, err
	}

	client := b.MailgunFactory(config.Domain, config.ApiKey)

	if err = client.DeleteCredential(username.(string)); err != nil {
		return logical.ErrorResponse("Unable to create credentials in mailgun. Configure with valid credentials"), nil
	}

	return nil, nil
}

func (b *backend) generateCredentials(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	username, err := generateUsername()
	if err != nil {
		return nil, err
	}
	password, err := generatePassword()
	if err != nil {
		return nil, err
	}

	config, err := getConfig(ctx, req.Storage)
	if ok, response, err := handleGetConfigErrors(err, config); !ok {
		return response, err
	}
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

func generateUsername() (string, error) {
	userSuffix, err := base62.Random(5, true)
	if err != nil {
		return "", errwrap.Wrapf("Unable to create unique, random username: {{err}}", err)
	}
	userSuffix = strings.ToLower(userSuffix)
	username := fmt.Sprintf("%v.%v", vaultUserPrefix, userSuffix)
	return username, nil
}

func generatePassword() (string, error) {
	password, err := credsutil.RandomAlphaNumeric(32, false)
	if err != nil {
		return "", errwrap.Wrapf("Unable to create random password: {{err}}", err)
	}
	return password, nil
}

func handleGetConfigErrors(err error, config *config) (bool, *logical.Response, error) {
	if err != nil {
		return false, nil, errwrap.Wrapf("Unable to get configuration: {{err}}", err)
	}
	if config == nil {
		return false, logical.ErrorResponse("Mailgun plugin is not configured."), nil
	}
	return true, nil, nil
}

const pathCredentialsSyn = `Generate Mailgun SMTP username and password.`
const pathCredentialsDesc = `
This path will generate new SMTP username and password for Mailgun.
The Mailgun SMTP server has following settings:
Server: smtp.mailgun.org
Ports 25, 587, and 465 (SSL/TLS)
`
