package mgsecret

import (
	"context"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"strings"
)

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := Backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

type backend struct {
	*framework.Backend
}

func Backend() *backend {
	var b backend
	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"config",
			},
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
