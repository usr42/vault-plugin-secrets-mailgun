package mgsecret

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"time"
)

func pathConfig(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"api_key": {
				Type:        framework.TypeString,
				Description: `Required. Mailgun API Key`,
				Required:    true,
			},
			"domain": {
				Type:        framework.TypeString,
				Description: "Required. Domain to generate SMTP credentials for",
				Required:    true,
			},
			"ttl": {
				Type:        framework.TypeDurationSecond,
				Description: "The Time to live (TTL) of the generated credentials",
			},
			"max_ttl": {
				Type:        framework.TypeDurationSecond,
				Description: "The maximum Time to live (TTL) of the generated credentials",
			},
		},

		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
				Summary:  "Return the current Mailgun configuration.",
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
				Summary:  "Configure the Mailgun settings.",
			},
		},

		HelpSynopsis:    pathConfigHelpSyn,
		HelpDescription: pathConfigHelpDesc,
	}
}

func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	cfg, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"domain":  cfg.Domain,
			"ttl":     int64(cfg.TTL / time.Second),
			"max_ttl": int64(cfg.MaxTTL / time.Second),
		},
	}, nil
}

func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	cfg, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &config{}
	}

	apiKey, resp := GetFieldString("api_key", data)
	if resp != nil {
		return resp, nil
	}
	cfg.ApiKey = apiKey

	domain, resp := GetFieldString("domain", data)
	if resp != nil {
		return resp, nil
	}
	cfg.Domain = domain

	// Update token TTL.
	if ttlRaw, ok := data.GetOk("ttl"); ok {
		cfg.TTL = time.Duration(ttlRaw.(int)) * time.Second
	}

	// Update token max TTL.
	if maxTtlRaw, ok := data.GetOk("max_ttl"); ok {
		cfg.MaxTTL = time.Duration(maxTtlRaw.(int)) * time.Second
	}

	client := b.MailgunFactory(domain, apiKey)
	if !client.IsApiKeyValid() {
		return logical.ErrorResponse("'api_key' is not valid."), nil
	}
	if !client.IsDomainValid() {
		return logical.ErrorResponse("'domain' is not valid."), nil
	}

	entry, err := logical.StorageEntryJSON("config", cfg)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

type config struct {
	ApiKey string
	Domain string
	TTL    time.Duration
	MaxTTL time.Duration
}

func getConfig(ctx context.Context, s logical.Storage) (*config, error) {
	var cfg config
	cfgRaw, err := s.Get(ctx, "config")
	if err != nil {
		return nil, err
	}
	if cfgRaw == nil {
		return nil, nil
	}

	if err := cfgRaw.DecodeJSON(&cfg); err != nil {
		return nil, err
	}

	return &cfg, err
}

func GetFieldString(key string, fields *framework.FieldData) (string, *logical.Response) {
	valueRaw, ok := fields.GetOk(key)
	if !ok {
		return "", logical.ErrorResponse(fmt.Sprintf("Required field '%s' is not set.", key))
	}
	return valueRaw.(string), nil
}

const pathConfigHelpSyn = `
Configure the Mailgun backend.
`

const pathConfigHelpDesc = `
The Mailgun backend requires the Mailgun API key and the domain for SMTP 
credentials. This endpoint is used to configure those credentials as well as 
default values for the backend in general.
`
