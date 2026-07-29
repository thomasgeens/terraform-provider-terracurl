package provider

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestReadFilePayload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload.bin")
	want := []byte{0x50, 0x4b, 0x03, 0x04}
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := readFilePayload(path)
	if err != nil {
		t.Fatalf("readFilePayload: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReadFilePayloadMissing(t *testing.T) {
	_, err := readFilePayload(filepath.Join(t.TempDir(), "missing.bin"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildMultipartBodyTextOnly(t *testing.T) {
	config := multipartConfigFromParts(t, []map[string]string{
		{"name": "name", "value": "John"},
	})
	body, contentType, diags := buildMultipartBody(config)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
		t.Fatalf("content type got %q", contentType)
	}
	if !strings.Contains(string(body), `name="name"`) || !strings.Contains(string(body), "John") {
		t.Fatalf("body missing text part: %q", body)
	}
}

func TestBuildMultipartBodyWithFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "photo.jpg")
	if err := os.WriteFile(filePath, []byte("image-bytes"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	config := multipartConfigFromParts(t, []map[string]string{
		{"name": "photo", "file_path": filePath, "content_type": "image/jpeg"},
	})
	body, contentType, diags := buildMultipartBody(config)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
		t.Fatalf("content type got %q", contentType)
	}
	if !strings.Contains(string(body), `filename="photo.jpg"`) || !strings.Contains(string(body), "image-bytes") {
		t.Fatalf("body missing file part: %q", body)
	}
}

func TestBuildMultipartBodyMixed(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "upload.txt")
	if err := os.WriteFile(filePath, []byte("file-content"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	config := multipartConfigFromParts(t, []map[string]string{
		{"name": "title", "value": "Example"},
		{"name": "attachment", "file_path": filePath},
	})
	_, _, diags := buildMultipartBody(config)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
}

func TestValidateMultipartPartsConflicts(t *testing.T) {
	parts := []MultipartPartModel{
		{
			Name:     types.StringValue("field"),
			Value:    types.StringValue("text"),
			FilePath: types.StringValue("/tmp/file"),
		},
	}
	diags := validateMultipartParts(parts, "")
	if !diags.HasError() {
		t.Fatal("expected validation error when value and file_path are both set")
	}
}

func TestResolveRequestPayloadEmpty(t *testing.T) {
	payload, diags := resolveRequestPayload(types.StringNull(), types.StringNull(), types.StringNull(), nil)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if len(payload.Body) != 0 {
		t.Fatalf("expected empty body, got %q", payload.Body)
	}
}

func TestResolveRequestPayloadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.txt")
	if err := os.WriteFile(path, []byte("from-disk"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	payload, diags := resolveRequestPayload(types.StringNull(), types.StringNull(), types.StringValue(path), nil)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if string(payload.Body) != "from-disk" || !payload.UsedFileBody {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestApplyResolvedPayloadToRequestOverridesContentType(t *testing.T) {
	body, contentType, diags := buildMultipartBody(multipartConfigFromParts(t, []map[string]string{
		{"name": "name", "value": "John"},
	}))
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	req := httptest.NewRequest(http.MethodPost, "https://example.com/upload", nil)
	req.Header.Set("Content-Type", "application/json")
	applyResolvedPayloadToRequest(req, resolvedRequestPayload{
		Body:              body,
		ContentType:       contentType,
		UsedMultipartBody: true,
	})

	if got := req.Header.Get("Content-Type"); got != contentType {
		t.Fatalf("content type got %q want %q", got, contentType)
	}
	if req.ContentLength != int64(len(body)) {
		t.Fatalf("content length got %d want %d", req.ContentLength, len(body))
	}
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(bodyBytes) != string(body) {
		t.Fatalf("body mismatch")
	}
}

func multipartConfigFromParts(t *testing.T, parts []map[string]string) *MultipartConfigModel {
	t.Helper()

	values := make([]attr.Value, 0, len(parts))
	for _, part := range parts {
		obj, diags := types.ObjectValue(multipartPartAttrTypes, map[string]attr.Value{
			"name":         types.StringValue(part["name"]),
			"value":        types.StringValue(part["value"]),
			"file_path":    types.StringValue(part["file_path"]),
			"content_type": types.StringValue(part["content_type"]),
		})
		if diags.HasError() {
			t.Fatalf("build part object: %v", diags)
		}
		values = append(values, obj)
	}

	list, diags := types.ListValue(multipartObjectType(), values)
	if diags.HasError() {
		t.Fatalf("build parts list: %v", diags)
	}
	return &MultipartConfigModel{Parts: list}
}
