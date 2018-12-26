package mgsecret

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/logical"
	"testing"
)

func TestPathCredentials(t *testing.T) {
	// TODO Add real tests
	t.Run("playground test", func(t *testing.T) {
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

		jsonpp(resp)
	})
}

func jsonpp(resp *logical.Response) {
	bytes, _ := json.MarshalIndent(resp.Secret, "", "  ")
	fmt.Println(string(bytes))
}
