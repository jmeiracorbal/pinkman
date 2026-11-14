package openai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmeiracorbal/pinkman"
)

func TestDiscoverListsOnlyEndpointModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"model-b"},{"id":"model-a"}]}`))
	}))
	defer server.Close()

	backend, err := New(Config{ID: "test", BaseURL: server.URL + "/v1", APIKey: "secret", Capabilities: []pinkman.Capability{pinkman.CapabilityText}})
	if err != nil {
		t.Fatal(err)
	}
	targets, err := backend.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 2 || targets[0].Model != "model-b" || targets[1].Model != "model-a" {
		t.Fatalf("unexpected targets: %#v", targets)
	}
}
