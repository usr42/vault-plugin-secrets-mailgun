package mgsecret

import (
	"context"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"strings"
)

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

type mailgunBackend struct {
	*framework.Backend
	MailgunFactory func(domain, apiKey string) MailgunClient
}

func backend() *mailgunBackend {
	var b mailgunBackend
	b.MailgunFactory = DefaultMailgunClientFactory
	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"config",
			},
		},
		Paths: framework.PathAppend(
			[]*framework.Path{
				pathConfig(&b),
				pathCredentials(&b),
			},
		),
		Secrets: []*framework.Secret{
			secretCredentials(&b),
		},
		BackendType: logical.TypeLogical,
	}
	return &b
}

const backendHelp = `
The Mailgun secrets backend dynamically generates Mailgun SMTP credentials 
for a given domain.The service account keys have a configurable lease set and 
are automatically revoked at the end of the lease.

After mounting this backend, credentials to generate Mailgun SMTP credentials 
must be configured with the "config/" endpoints and policies must be
written using the "roles/" endpoints before any keys can be generated.
`
