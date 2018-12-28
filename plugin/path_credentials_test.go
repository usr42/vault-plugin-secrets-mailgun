package mgsecret

import (
	"context"
	"github.com/hashicorp/vault/logical"
	"testing"
	"time"
)

func TestPathCredentials(t *testing.T) {
	t.Run("not configured plugin does not provide credentials", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Storage:   storage,
			Operation: logical.ReadOperation,
			Path:      "credentials",
		})
		if err != nil {
			t.Fatal(err)
		}

		if !resp.IsError() {
			t.Error("Not configured plugin still created credentials:", resp.Secret)
		}
	})

	t.Run("configured plugin provides credential", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		storeDefaultConfig(t, b, storage)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Storage:   storage,
			Operation: logical.ReadOperation,
			Path:      "credentials",
		})
		if err != nil {
			t.Fatal(err)
		}

		if resp.IsError() {
			t.Error("Configured plugin is an error but should not")
		}

		data := resp.Data
		if _, ok := data["username"]; !ok {
			t.Error("Response does not contain username")
		}
		if _, ok := data["password"]; !ok {
			t.Error("Response does not contain password")
		}
	})

	t.Run("credentials have configured ttl and max_ttl", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		expectedTTLString := "1h"
		expectedMaxTTLString := "6h"
		expectedTTL, err := time.ParseDuration(expectedTTLString)
		if err != nil {
			t.Fatal("Cannot convert duration")
		}
		expectedMaxTTL, err := time.ParseDuration(expectedMaxTTLString)
		if err != nil {
			t.Fatal("Cannot convert duration")
		}
		config := map[string]interface{}{
			"api_key": "apiKey123",
			"domain":  "example.com",
			"ttl":     expectedTTLString,
			"max_ttl": expectedMaxTTLString,
		}
		storeConfig(config, t, b, storage)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Storage:   storage,
			Operation: logical.ReadOperation,
			Path:      "credentials",
		})
		if err != nil {
			t.Fatal(err)
		}

		if resp.IsError() {
			t.Error("Configured plugin is an error but should not")
		}

		secret := resp.Secret
		if secret.TTL != expectedTTL {
			t.Error("Credential ttl should be", expectedTTL, "but is", secret.TTL)
		}
		if secret.MaxTTL != expectedMaxTTL {
			t.Error("Credential ttl should be", expectedMaxTTL, "but is", secret.MaxTTL)
		}
	})
}
