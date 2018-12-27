package mgsecret

import (
	"context"
	"github.com/hashicorp/vault/logical"
	"testing"
	"time"
)

func TestPathConfig(t *testing.T) {
	t.Run("initial config is empty", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)

		resp := requestConfig(t, b, storage)

		if resp != nil {
			t.Error("Unexpected intial config:", resp)
		}
	})

	t.Run("saving only domain is not allowed", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		config := map[string]interface{}{
			"domain": "example.com",
		}

		response := storeConfig(config, t, b, storage)

		if !response.IsError() {
			t.Error("Saving only domain did not results in error response")
		}
		resp := requestConfig(t, b, storage)
		if resp != nil {
			t.Fatal("Configuration without api_key was still saved.")
		}
	})

	t.Run("saving only api_key is not allowed", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		config := map[string]interface{}{
			"api_key": "apiKey123",
		}
		response := storeConfig(config, t, b, storage)

		if !response.IsError() {
			t.Error("Saving only api_key did not results in error response")
		}

		resp := requestConfig(t, b, storage)

		if resp != nil {
			t.Fatal("Configuration without domain was still saved.")
		}
	})

	t.Run("api_key cannot be read", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		config := map[string]interface{}{
			"domain":  "example.com",
			"api_key": "apiKey123",
		}
		storeConfig(config, t, b, storage)

		resp := requestConfig(t, b, storage)

		if resp == nil {
			t.Fatal("configuration", config, "was not saved.")
		}
		if apiKey, ok := resp.Data["api_key"]; ok {
			t.Error("api_key must not be readable, but was:", apiKey)
		}
	})

	t.Run("saving config, contains domain", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		expectedDomain := "example.com"
		config := map[string]interface{}{
			"api_key": "apiKey123",
			"domain":  expectedDomain,
		}
		storeConfig(config, t, b, storage)

		resp := requestConfig(t, b, storage)

		if resp == nil {
			t.Fatal("configuration", config, "was not saved.")
		}
		data := resp.Data
		domain, ok := data["domain"]
		if !ok {
			t.Error("domain", expectedDomain, "was not saved")
		}
		if domain != expectedDomain {
			t.Error("domain was", domain, ", but expected", expectedDomain)
		}
		if apiKey, ok := data["api_key"]; ok {
			t.Error("api_key must not be readable, but was:", apiKey)
		}
	})

	t.Run("api_key is not readable in config", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		config := storeDefaultConfig(t, b, storage)

		resp := requestConfig(t, b, storage)

		if resp == nil {
			t.Fatal("configuration", config, "was not saved.")
		}

		data := resp.Data

		if apiKey, ok := data["api_key"]; ok {
			t.Error("api_key must not be readable, but was:", apiKey)
		}
	})

	t.Run("config contains ttl", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		config := storeDefaultConfig(t, b, storage)

		resp := requestConfig(t, b, storage)

		if resp == nil {
			t.Fatal("configuration", config, "was not saved.")
		}

		data := resp.Data

		ttl, ok := data["ttl"]
		if !ok {
			t.Fatal("configuration does not contain ttl")
		}

		expectedTTL := 10 * time.Minute
		if ttl != expectedTTL {
			t.Error("ttl is not the default value. Expected:", expectedTTL, "Actual:", ttl)
		}
	})

	t.Run("set ttl", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		config := map[string]interface{}{
			"api_key": "apiKey123",
			"domain":  "example.com",
			"ttl":     "1h",
		}
		storeConfig(config, t, b, storage)

		resp := requestConfig(t, b, storage)

		if resp == nil {
			t.Fatal("configuration", config, "was not saved.")
		}

		data := resp.Data

		ttl, ok := data["ttl"]
		if !ok {
			t.Fatal("configuration does not contain ttl")
		}

		expectedTTL := 1 * time.Hour
		if ttl != expectedTTL {
			t.Error("ttl is not the default value. Expected:", expectedTTL, "Actual:", ttl)
		}
	})
}

var defaultConfig = map[string]interface{}{
	"api_key": "apiKey123",
	"domain":  "example.com",
}

func storeDefaultConfig(t *testing.T, b *backend, storage logical.Storage) map[string]interface{} {
	config := defaultConfig
	storeConfig(config, t, b, storage)
	return config
}

func testBackend(tb testing.TB) (*backend, logical.Storage) {
	tb.Helper()

	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}

	b, err := Factory(context.Background(), config)
	if err != nil {
		tb.Fatal(err)
	}
	backend := b.(*backend)
	backend.MailgunFactory = testMailgunClientFactory
	return backend, config.StorageView
}

func requestConfig(t *testing.T, b *backend, storage logical.Storage) *logical.Response {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Storage:   storage,
		Operation: logical.ReadOperation,
		Path:      "config",
	})
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func storeConfig(config map[string]interface{}, t *testing.T, b *backend, storage logical.Storage) *logical.Response {
	response, err := b.HandleRequest(context.Background(), &logical.Request{
		Storage:   storage,
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data:      config,
	})
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func testMailgunClientFactory(_, _ string) MailgunClient {
	return testMailgunClient{}
}

type testMailgunClient struct{}

func (c testMailgunClient) IsDomainValid() bool {
	return true
}

func (c testMailgunClient) IsApiKeyValid() bool {
	return true
}

func (c testMailgunClient) DeleteCredential(username string) error {
	return nil
}

func (c testMailgunClient) CreateCredential(login, password string) error {
	return nil
}
