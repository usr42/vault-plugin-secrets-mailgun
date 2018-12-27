package mgsecret

import (
	"context"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"time"
)

const (
	defaultTTL = "10m"
)

func pathConfig(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"api_key": {
				Type:        framework.TypeString,
				Description: `Required. Mailgun API Key`,
			},
			"domain": {
				Type:        framework.TypeString,
				Description: "Required. Domain to generate SMTP credentials for",
			},
			"ttl": {
				Type:        framework.TypeDurationSecond,
				Description: "The Time to live (TTL) of the generated credentials",
				Default:     defaultTTL,
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
			"domain": cfg.Domain,
			"ttl":    cfg.TTL,
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

	apiKeyRaw, ok := data.GetOk("api_key")
	if !ok {
		return logical.ErrorResponse("Required field 'api_key' is not set."), nil
	}
	apiKey := apiKeyRaw.(string)
	cfg.ApiKey = apiKey

	domainRaw, ok := data.GetOk("domain")
	if !ok {
		return logical.ErrorResponse("Required field 'domain' is not set."), nil
	}
	domain := domainRaw.(string)
	cfg.Domain = domain

	ttlRaw := data.Get("ttl")
	cfg.TTL = time.Duration(ttlRaw.(int)) * time.Second

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

const pathConfigHelpSyn = `
Configure the Mailgun backend.
`

const pathConfigHelpDesc = `
The Mailgun backend requires the Mailgun API key and the domain for SMTP 
credentials. This endpoint is used to configure those credentials as well as 
default values for the backend in general.
`
