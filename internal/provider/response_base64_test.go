package provider

import (
	"encoding/base64"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEncodeResponseBase64(t *testing.T) {
	want := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	got := encodeResponseBase64(want)
	if got != "iVBORw0KGgo=" {
		t.Fatalf("got %q", got)
	}

	if encodeResponseBase64(nil) != "" {
		t.Fatal("expected empty string for nil body")
	}
	if encodeResponseBase64([]byte{}) != "" {
		t.Fatal("expected empty string for empty body")
	}
}

func TestSetDataSourceResponseValuesFromBytes(t *testing.T) {
	body := []byte{0x00, 0x01, 0x02}
	encoded := base64.StdEncoding.EncodeToString(body)

	t.Run("non-sensitive", func(t *testing.T) {
		data := &CurlDataSourceModel{
			ResponseSensitive: types.BoolValue(false),
		}
		setDataSourceResponseValuesFromBytes(data, body, string(body))

		if data.Response.ValueString() != string(body) {
			t.Fatalf("response got %q", data.Response.ValueString())
		}
		if data.ResponseBase64.ValueString() != encoded {
			t.Fatalf("response_base64 got %q want %q", data.ResponseBase64.ValueString(), encoded)
		}
		if data.SensitiveResponse.ValueString() != "" || data.SensitiveResponseBase64.ValueString() != "" {
			t.Fatal("expected sensitive response attrs to be empty")
		}
	})

	t.Run("sensitive", func(t *testing.T) {
		data := &CurlDataSourceModel{
			ResponseSensitive: types.BoolValue(true),
		}
		setDataSourceResponseValuesFromBytes(data, body, string(body))

		if data.Response.ValueString() != "" || data.ResponseBase64.ValueString() != "" {
			t.Fatal("expected non-sensitive response attrs to be empty")
		}
		if data.SensitiveResponse.ValueString() != string(body) {
			t.Fatalf("sensitive_response got %q", data.SensitiveResponse.ValueString())
		}
		if data.SensitiveResponseBase64.ValueString() != encoded {
			t.Fatalf("sensitive_response_base64 got %q want %q", data.SensitiveResponseBase64.ValueString(), encoded)
		}
	})
}

func TestResponseBodyContainsInvalidUTF8(t *testing.T) {
	if !responseBodyContainsInvalidUTF8([]byte{0xff, 0xfe, 0xfd}) {
		t.Fatal("expected invalid UTF-8 detection")
	}
	if responseBodyContainsInvalidUTF8([]byte("hello")) {
		t.Fatal("expected valid UTF-8")
	}
	if responseBodyContainsInvalidUTF8(nil) {
		t.Fatal("expected empty body to be valid")
	}
}
