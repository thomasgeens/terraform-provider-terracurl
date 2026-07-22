package provider

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNewProviderMetaDefaultHeaders(t *testing.T) {
	headers := types.MapValueMust(types.StringType, map[string]attr.Value{
		"Authorization": types.StringValue("Bearer fresh"),
		"X-Custom":      types.StringValue("value"),
	})

	meta := NewProviderMeta(
		types.StringNull(),
		types.StringNull(),
		types.StringNull(),
		headers,
	)

	got := meta.DefaultHeaders()
	if got["Authorization"] != "Bearer fresh" {
		t.Fatalf("expected Authorization header, got %q", got["Authorization"])
	}
	if got["X-Custom"] != "value" {
		t.Fatalf("expected X-Custom header, got %q", got["X-Custom"])
	}
}

func TestNewProviderMetaNullDefaultHeaders(t *testing.T) {
	meta := NewProviderMeta(
		types.StringNull(),
		types.StringNull(),
		types.StringNull(),
		types.MapNull(types.StringType),
	)

	if meta.DefaultHeaders() != nil {
		t.Fatalf("expected nil default headers, got %v", meta.DefaultHeaders())
	}
}

func TestApplyRequestHeadersWithDefaultsOverridesResourceHeaders(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, "https://example.com/resource", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resourceHeaders := types.MapValueMust(types.StringType, map[string]attr.Value{
		"Authorization": types.StringValue("Bearer stale"),
		"Content-Type":  types.StringValue("application/json"),
	})
	providerHeaders := types.MapValueMust(types.StringType, map[string]attr.Value{
		"Authorization": types.StringValue("Bearer fresh"),
	})
	meta := NewProviderMeta(
		types.StringNull(),
		types.StringNull(),
		types.StringNull(),
		providerHeaders,
	)

	applyRequestHeadersWithDefaults(req, resourceHeaders, meta)

	if got := req.Header.Get("Authorization"); got != "Bearer fresh" {
		t.Fatalf("expected provider Authorization to override stale value, got %q", got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type from resource headers, got %q", got)
	}
}

func TestApplyRequestHeadersWithDefaultsHostOverride(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/path", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resourceHeaders := types.MapValueMust(types.StringType, map[string]attr.Value{
		"host": types.StringValue("stale.example.com"),
	})
	providerHeaders := types.MapValueMust(types.StringType, map[string]attr.Value{
		"Host": types.StringValue("fresh.example.com"),
	})
	meta := NewProviderMeta(
		types.StringNull(),
		types.StringNull(),
		types.StringNull(),
		providerHeaders,
	)

	applyRequestHeadersWithDefaults(req, resourceHeaders, meta)

	if req.Host != "fresh.example.com" {
		t.Fatalf("expected provider Host override, got %q", req.Host)
	}
}

func TestApplyRequestHeadersWithDefaultsNilMeta(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resourceHeaders := types.MapValueMust(types.StringType, map[string]attr.Value{
		"Authorization": types.StringValue("Bearer stale"),
	})

	applyRequestHeadersWithDefaults(req, resourceHeaders, nil)

	if got := req.Header.Get("Authorization"); got != "Bearer stale" {
		t.Fatalf("expected resource Authorization unchanged, got %q", got)
	}
}

func TestDefaultProviderMetaHasNoDefaultHeaders(t *testing.T) {
	meta := DefaultProviderMeta()
	if meta.DefaultHeaders() != nil {
		t.Fatalf("expected no default headers, got %v", meta.DefaultHeaders())
	}
}
