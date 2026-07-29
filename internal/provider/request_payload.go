package provider

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resolvedRequestPayload struct {
	Body              []byte
	ContentType       string
	UsedWriteOnlyBody bool
	UsedFileBody      bool
	UsedMultipartBody bool
}

func resolveRequestPayload(regular, writeOnly, filePath types.String, multipart *MultipartConfigModel) (resolvedRequestPayload, diag.Diagnostics) {
	var result resolvedRequestPayload
	var diags diag.Diagnostics

	if multipartConfigIsSet(multipart) {
		body, contentType, buildDiags := buildMultipartBody(multipart)
		diags.Append(buildDiags...)
		if diags.HasError() {
			return result, diags
		}
		result.Body = body
		result.ContentType = contentType
		result.UsedMultipartBody = true
		return result, diags
	}

	if hasValue(filePath) {
		body, err := readFilePayload(filePath.ValueString())
		if err != nil {
			diags.AddError("Request Body File Error", err.Error())
			return result, diags
		}
		result.Body = body
		result.UsedFileBody = true
		return result, diags
	}

	body, usedWriteOnly := resolveRequestBody(regular, writeOnly)
	result.Body = body
	result.UsedWriteOnlyBody = usedWriteOnly
	return result, diags
}

func readFilePayload(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("file path is empty")
	}
	body, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read request body file %q: %w", path, err)
	}
	return body, nil
}

func buildMultipartBody(config *MultipartConfigModel) ([]byte, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	parts, parseDiags := parseMultipartParts(config.Parts)
	diags.Append(parseDiags...)
	diags.Append(validateMultipartParts(parts, "")...)
	if diags.HasError() {
		return nil, "", diags
	}

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	for i, part := range parts {
		fieldPrefix := fmt.Sprintf("parts[%d]", i)
		if err := writeMultipartPart(writer, part, fieldPrefix); err != nil {
			diags.AddError("Multipart Configuration Error", err.Error())
			return nil, "", diags
		}
	}

	if err := writer.Close(); err != nil {
		diags.AddError("Multipart Configuration Error", fmt.Sprintf("failed to finalize multipart body: %s", err))
		return nil, "", diags
	}

	return buffer.Bytes(), writer.FormDataContentType(), diags
}

func writeMultipartPart(writer *multipart.Writer, part MultipartPartModel, fieldPrefix string) error {
	partContentType := part.ContentType.ValueString()

	if hasValue(part.Value) {
		contentType := partContentType
		if contentType == "" {
			contentType = "text/plain"
		}
		header := textPartHeader(part.Name.ValueString(), contentType)
		formPart, err := writer.CreatePart(header)
		if err != nil {
			return fmt.Errorf("%s: failed to create text part: %w", fieldPrefix, err)
		}
		if _, err := io.WriteString(formPart, part.Value.ValueString()); err != nil {
			return fmt.Errorf("%s: failed to write text part: %w", fieldPrefix, err)
		}
		return nil
	}

	filePath := part.FilePath.ValueString()
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("%s: failed to open file %q: %w", fieldPrefix, filePath, err)
	}
	defer func() { _ = file.Close() }()

	contentType := partContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	formPart, err := writer.CreatePart(filePartHeader(part.Name.ValueString(), filepath.Base(filePath), contentType))
	if err != nil {
		return fmt.Errorf("%s: failed to create file part: %w", fieldPrefix, err)
	}
	if _, err := io.Copy(formPart, file); err != nil {
		return fmt.Errorf("%s: failed to write file part: %w", fieldPrefix, err)
	}
	return nil
}

func textPartHeader(name, contentType string) map[string][]string {
	return map[string][]string{
		"Content-Disposition": {fmt.Sprintf(`form-data; name=%q`, name)},
		"Content-Type":        {contentType},
	}
}

func filePartHeader(name, filename, contentType string) map[string][]string {
	return map[string][]string{
		"Content-Disposition": {fmt.Sprintf(`form-data; name=%q; filename=%q`, name, filename)},
		"Content-Type":        {contentType},
	}
}

func applyResolvedPayloadToRequest(request *http.Request, payload resolvedRequestPayload) {
	if len(payload.Body) > 0 {
		request.Body = io.NopCloser(bytes.NewReader(payload.Body))
		request.ContentLength = int64(len(payload.Body))
	}
	if payload.ContentType != "" {
		request.Header.Set("Content-Type", payload.ContentType)
	}
}

func payloadLogLabel(payload resolvedRequestPayload, regular types.String) string {
	switch {
	case payload.UsedMultipartBody:
		return "<multipart request body>"
	case payload.UsedFileBody:
		return "<file request body>"
	case payload.UsedWriteOnlyBody:
		return "<redacted write-only request body>"
	case hasValue(regular):
		return regular.ValueString()
	default:
		return ""
	}
}
