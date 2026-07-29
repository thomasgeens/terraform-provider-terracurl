package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type fakePrivateState struct {
	data map[string][]byte
}

func (f *fakePrivateState) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if f.data == nil {
		return nil, nil
	}
	return f.data[key], nil
}

func (f *fakePrivateState) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	if f.data == nil {
		f.data = make(map[string][]byte)
	}
	if len(value) == 0 {
		delete(f.data, key)
		return nil
	}
	f.data[key] = value
	return nil
}

func TestResolveRequestBodyPrefersWriteOnly(t *testing.T) {
	regular := types.StringValue(`{"stored":true}`)
	writeOnly := types.StringValue(`{"secret":true}`)

	body, usedWo := resolveRequestBody(regular, writeOnly)
	if !usedWo || string(body) != `{"secret":true}` {
		t.Fatalf("expected write-only body, got %q usedWo=%v", body, usedWo)
	}
}

func TestResolveRequestBodyUsesRegularWhenWriteOnlyUnset(t *testing.T) {
	regular := types.StringValue(`{"stored":true}`)
	body, usedWo := resolveRequestBody(regular, types.StringNull())
	if usedWo || string(body) != `{"stored":true}` {
		t.Fatalf("expected regular body, got %q usedWo=%v", body, usedWo)
	}
}

func TestApplyRequestHeadersWithWriteOnlyOverridesDefaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	headers := types.MapValueMust(types.StringType, map[string]attr.Value{
		"X-Resource": types.StringValue("resource"),
	})
	headersWo := types.MapValueMust(types.StringType, map[string]attr.Value{
		"Authorization": types.StringValue("Bearer write-only"),
		"X-Resource":    types.StringValue("override"),
	})
	meta := NewProviderMeta(
		types.StringNull(),
		types.StringNull(),
		types.StringNull(),
		types.MapValueMust(types.StringType, map[string]attr.Value{
			"Authorization": types.StringValue("Bearer provider"),
		}),
		nil,
	)

	applyRequestHeadersWithWriteOnly(req, headers, headersWo, meta)

	if got := req.Header.Get("Authorization"); got != "Bearer write-only" {
		t.Fatalf("expected write-only Authorization, got %q", got)
	}
	if got := req.Header.Get("X-Resource"); got != "override" {
		t.Fatalf("expected write-only header override, got %q", got)
	}
}

func TestWriteOnlyPrivateSnapshotRoundtrip(t *testing.T) {
	ctx := context.Background()
	private := &fakePrivateState{}
	config := CurlResourceModel{
		HeadersWo: types.MapValueMust(types.StringType, map[string]attr.Value{
			"Authorization": types.StringValue("Bearer create"),
		}),
		ReadHeadersWo: types.MapValueMust(types.StringType, map[string]attr.Value{
			"Authorization": types.StringValue("Bearer read"),
		}),
		DestroyHeadersWo: types.MapValueMust(types.StringType, map[string]attr.Value{
			"Authorization": types.StringValue("Bearer destroy"),
		}),
		RequestBodyWo:        types.StringValue(`{"create":true}`),
		ReadRequestBodyWo:    types.StringValue(`{"read":true}`),
		DestroyRequestBodyWo: types.StringValue(`{"destroy":true}`),
	}

	if diags := snapshotWriteOnlyToPrivate(ctx, config, private); diags.HasError() {
		t.Fatalf("snapshot failed: %v", diags)
	}

	var loaded CurlResourceModel
	if diags := loadWriteOnlyFromPrivate(ctx, private, &loaded); diags.HasError() {
		t.Fatalf("load failed: %v", diags)
	}

	auth, ok := loaded.DestroyHeadersWo.Elements()["Authorization"].(types.String)
	if !ok || auth.ValueString() != "Bearer destroy" {
		t.Fatalf("unexpected destroy headers %#v", loaded.DestroyHeadersWo)
	}
	if loaded.ReadRequestBodyWo.ValueString() != `{"read":true}` {
		t.Fatalf("unexpected read body %q", loaded.ReadRequestBodyWo.ValueString())
	}
}

func TestDestroyWriteOnlyLoadedFromPrivateStateAppliedToRequest(t *testing.T) {
	ctx := context.Background()
	private := &fakePrivateState{}
	config := CurlResourceModel{
		DestroyHeadersWo: types.MapValueMust(types.StringType, map[string]attr.Value{
			"Authorization": types.StringValue("Bearer destroy-secret"),
		}),
	}
	if diags := snapshotWriteOnlyToPrivate(ctx, config, private); diags.HasError() {
		t.Fatalf("snapshot failed: %v", diags)
	}

	var data CurlResourceModel
	if diags := loadWriteOnlyFromPrivate(ctx, private, &data); diags.HasError() {
		t.Fatalf("load failed: %v", diags)
	}

	req := httptest.NewRequest(http.MethodDelete, "https://example.com/destroy", nil)
	applyRequestHeadersWithWriteOnly(req, types.MapNull(types.StringType), data.DestroyHeadersWo, nil)
	if got := req.Header.Get("Authorization"); got != "Bearer destroy-secret" {
		t.Fatalf("expected destroy write-only header on request, got %q", got)
	}
}

func TestWriteOnlyVersionsChanged(t *testing.T) {
	state := CurlResourceModel{HeadersWoVersion: types.Int64Value(1)}
	config := CurlResourceModel{HeadersWoVersion: types.Int64Value(2)}
	if !writeOnlyVersionsChanged(state, config) {
		t.Fatal("expected version change to be detected")
	}
}

func TestRequestBodyForLogRedactsWriteOnly(t *testing.T) {
	got := requestBodyForLog(types.StringNull(), true)
	if got != "<redacted write-only request body>" {
		t.Fatalf("unexpected log value %q", got)
	}
}
