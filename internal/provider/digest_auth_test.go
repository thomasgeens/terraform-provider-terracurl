package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	resource2 "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveDigestAuthPrefersOverride(t *testing.T) {
	providerAuth := &DigestAuthConfig{Username: "provider", Password: "provider-pass"}
	override := &DigestAuthModel{
		Username: types.StringValue("resource"),
		Password: types.StringValue("resource-pass"),
	}

	got := resolveDigestAuth(providerAuth, override)
	if got == nil || got.Username != "resource" || got.Password != "resource-pass" {
		t.Fatalf("expected resource override, got %#v", got)
	}
}

func TestResolveDigestAuthUsesProviderDefault(t *testing.T) {
	providerAuth := &DigestAuthConfig{Username: "provider", Password: "provider-pass"}
	got := resolveDigestAuth(providerAuth, nil)
	if got == nil || got.Username != "provider" {
		t.Fatalf("expected provider default, got %#v", got)
	}
}

func TestResolveDigestAuthNilWhenUnset(t *testing.T) {
	if got := resolveDigestAuth(nil, nil); got != nil {
		t.Fatalf("expected nil digest auth, got %#v", got)
	}
}

func TestNewHTTPClientWithDigestAuth(t *testing.T) {
	const (
		username = "testuser"
		password = "testpass"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Digest ") {
			w.Header().Set("WWW-Authenticate", `Digest realm="test", nonce="abc123", algorithm=MD5, qop="auth"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("authenticated"))
	}))
	defer server.Close()

	meta := NewProviderMeta(
		types.StringNull(),
		types.StringNull(),
		types.StringNull(),
		types.MapNull(types.StringType),
		&DigestAuthModel{
			Username: types.StringValue(username),
			Password: types.StringValue(password),
		},
	)

	client, err := meta.NewHTTPClient(nil, meta.DefaultDigestAuth())
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", resp.StatusCode, body)
	}
	if string(body) != "authenticated" {
		t.Fatalf("unexpected body %q", body)
	}
}

func TestDigestAuthFromPrivateMap(t *testing.T) {
	model := digestAuthFromPrivateMap(map[string]interface{}{
		"username": "stored",
		"password": "secret",
	})
	if model == nil || model.Username.ValueString() != "stored" || model.Password.ValueString() != "secret" {
		t.Fatalf("unexpected model %#v", model)
	}
}

func TestCurlResource_Delete_ProviderDefaultDigestAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Digest ") {
			w.Header().Set("WWW-Authenticate", `Digest realm="test", nonce="abc123", algorithm=MD5, qop="auth"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"deleted":true}`))
	}))
	defer server.Close()

	ctx := context.Background()
	meta := NewProviderMeta(
		types.StringNull(),
		types.StringNull(),
		types.StringNull(),
		types.MapNull(types.StringType),
		&DigestAuthModel{
			Username: types.StringValue("admin"),
			Password: types.StringValue("secret"),
		},
	)
	r := &CurlResource{meta: meta}

	schemaResp := &resource2.SchemaResponse{}
	r.Schema(ctx, resource2.SchemaRequest{}, schemaResp)

	stateModel := CurlResourceModel{
		Id:                   types.StringValue("test"),
		Name:                 types.StringValue("test"),
		Url:                  types.StringValue(server.URL + "/create"),
		Method:               types.StringValue("POST"),
		SkipRead:             types.BoolValue(true),
		SkipDestroy:          types.BoolValue(false),
		DestroyUrl:           types.StringValue(server.URL + "/destroy"),
		DestroyMethod:        types.StringValue("DELETE"),
		DestroyResponseCodes: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200")}),
		DestroyTimeout:           types.Int64Value(10),
		DestroyRetryInterval:     types.Int64Value(1),
		DestroyMaxRetry:          types.Int64Value(0),
		ResponseCodes:            types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200")}),
		ReadResponseCodes:        types.ListNull(types.StringType),
		IgnoreResponseFields:     types.ListNull(types.StringType),
		Headers:                  types.MapNull(types.StringType),
		RequestParameters:        types.MapNull(types.StringType),
		ReadHeaders:              types.MapNull(types.StringType),
		ReadParameters:           types.MapNull(types.StringType),
		DestroyHeaders:           types.MapNull(types.StringType),
		DestroyRequestParameters: types.MapNull(types.StringType),
	}

	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(ctx, &stateModel); diags.HasError() {
		t.Fatalf("failed to set state: %v", diags)
	}

	deleteResp := &resource2.DeleteResponse{State: state}
	r.Delete(ctx, resource2.DeleteRequest{State: state}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete failed: %v", deleteResp.Diagnostics)
	}
}
