package mgsecret

import (
	"context"
	"github.com/hashicorp/vault/logical"
	"testing"
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

	t.Run("saving domain", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		expectedDomain := "example.com"
		config := map[string]interface{}{
			"domain": expectedDomain,
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
		if key, ok := data["api_key"]; ok {
			t.Error("'api_key' was unexpectedly received:", key)
		}
		if domain != expectedDomain {
			t.Error("domain was", domain, ", but expected", expectedDomain)
		}
	})

	t.Run("api_key cannot be read", func(t *testing.T) {
		t.Parallel()
		b, storage := testBackend(t)
		config := map[string]interface{}{
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

	t.Run("saving domain and api_key", func(t *testing.T) {
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
}

func testBackend(tb testing.TB) (*backend, logical.Storage) {
	tb.Helper()

	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}

	b, err := Factory(context.Background(), config)
	if err != nil {
		tb.Fatal(err)
	}
	return b.(*backend), config.StorageView
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

func storeConfig(config map[string]interface{}, t *testing.T, b *backend, storage logical.Storage) {
	_, err := b.HandleRequest(context.Background(), &logical.Request{
		Storage:   storage,
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data:      config,
	})
	if err != nil {
		t.Fatal(err)
	}
}
