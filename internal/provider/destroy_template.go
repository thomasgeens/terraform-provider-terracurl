package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const responsePlaceholderPrefix = "{response."

var responsePlaceholderPattern = regexp.MustCompile(`\{response\.([^}]+)\}`)

type resolvedDestroyTemplates struct {
	URL              string
	Payload          resolvedRequestPayload
	DestroyHeaders   types.Map
	DestroyHeadersWo types.Map
	DestroyParams    types.Map
}

func containsResponsePlaceholder(s string) bool {
	return strings.Contains(s, responsePlaceholderPrefix)
}

func mapContainsResponsePlaceholder(m types.Map) bool {
	if m.IsNull() || m.IsUnknown() {
		return false
	}
	for _, v := range m.Elements() {
		if strVal, ok := v.(types.String); ok && containsResponsePlaceholder(strVal.ValueString()) {
			return true
		}
	}
	return false
}

func destroyFieldsHavePlaceholders(data *CurlResourceModel) bool {
	if data == nil {
		return false
	}
	if hasValue(data.DestroyUrl) && containsResponsePlaceholder(data.DestroyUrl.ValueString()) {
		return true
	}
	if hasValue(data.DestroyRequestBody) && containsResponsePlaceholder(data.DestroyRequestBody.ValueString()) {
		return true
	}
	if hasValue(data.DestroyRequestBodyWo) && containsResponsePlaceholder(data.DestroyRequestBodyWo.ValueString()) {
		return true
	}
	if hasValue(data.DestroyRequestBodyFile) && containsResponsePlaceholder(data.DestroyRequestBodyFile.ValueString()) {
		return true
	}
	if multipartConfigHasPlaceholders(data.DestroyRequestMultipart) {
		return true
	}
	return mapContainsResponsePlaceholder(data.DestroyHeaders) ||
		mapContainsResponsePlaceholder(data.DestroyHeadersWo) ||
		mapContainsResponsePlaceholder(data.DestroyRequestParameters)
}

func extractJSONPath(jsonStr, path string) (string, error) {
	if jsonStr == "" {
		return "", fmt.Errorf("response is empty")
	}

	var root interface{}
	if err := json.Unmarshal([]byte(jsonStr), &root); err != nil {
		return "", fmt.Errorf("response is not valid JSON: %w", err)
	}

	current := root
	for _, seg := range strings.Split(path, ".") {
		if current == nil {
			return "", fmt.Errorf("path %q not found in response", path)
		}

		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[seg]
			if !ok {
				return "", fmt.Errorf("path %q not found in response", path)
			}
			current = val
		case []interface{}:
			idx, err := strconv.Atoi(seg)
			if err != nil {
				return "", fmt.Errorf("path %q: expected array index, got %q", path, seg)
			}
			if idx < 0 || idx >= len(v) {
				return "", fmt.Errorf("path %q not found in response", path)
			}
			current = v[idx]
		default:
			return "", fmt.Errorf("path %q not found in response", path)
		}
	}

	return stringifyJSONValue(current)
}

func stringifyJSONValue(v interface{}) (string, error) {
	if v == nil {
		return "", fmt.Errorf("value at path is null")
	}

	switch val := v.(type) {
	case string:
		return val, nil
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10), nil
		}
		return strconv.FormatFloat(val, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(val), nil
	default:
		return "", fmt.Errorf("value at path is not a scalar (got %T)", v)
	}
}

func unescapeDoubleBraces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if i+1 < len(s) && s[i] == '{' && s[i+1] == '{' {
			b.WriteByte('{')
			i++
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

var escapedResponsePlaceholderPattern = regexp.MustCompile(`\{\{response\.([^}]+)\}\}`)

func substituteResponsePlaceholders(template, responseJSON string) (string, error) {
	if !containsResponsePlaceholder(template) {
		return template, nil
	}

	type escapedToken struct {
		literal string
	}
	var tokens []escapedToken

	protected := escapedResponsePlaceholderPattern.ReplaceAllStringFunc(template, func(match string) string {
		submatches := escapedResponsePlaceholderPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		idx := len(tokens)
		tokens = append(tokens, escapedToken{literal: "{response." + submatches[1] + "}"})
		return fmt.Sprintf("\x00ESC%d\x00", idx)
	})

	unescaped := unescapeDoubleBraces(protected)

	var substErr error
	result := responsePlaceholderPattern.ReplaceAllStringFunc(unescaped, func(match string) string {
		if substErr != nil {
			return match
		}
		submatches := responsePlaceholderPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		value, err := extractJSONPath(responseJSON, submatches[1])
		if err != nil {
			substErr = fmt.Errorf("placeholder %s: %w", match, err)
			return match
		}
		return value
	})
	if substErr != nil {
		return "", substErr
	}

	for i, token := range tokens {
		result = strings.ReplaceAll(result, fmt.Sprintf("\x00ESC%d\x00", i), token.literal)
	}

	return result, nil
}

func substituteFieldPlaceholders(value, responseJSON, fieldName string) (string, error) {
	if !containsResponsePlaceholder(value) {
		return value, nil
	}
	result, err := substituteResponsePlaceholders(value, responseJSON)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fieldName, err)
	}
	return result, nil
}

func substituteMapPlaceholders(m types.Map, responseJSON, fieldName string) (types.Map, error) {
	if m.IsNull() || m.IsUnknown() {
		return m, nil
	}

	result := make(map[string]types.String, len(m.Elements()))
	for k, v := range m.Elements() {
		strVal, ok := v.(types.String)
		if !ok {
			continue
		}
		substituted, err := substituteFieldPlaceholders(strVal.ValueString(), responseJSON, fieldName)
		if err != nil {
			return types.MapNull(types.StringType), err
		}
		result[k] = types.StringValue(substituted)
	}

	tfMap, diags := types.MapValueFrom(context.Background(), types.StringType, result)
	if diags.HasError() {
		return types.MapNull(types.StringType), fmt.Errorf("%s: failed to build map value", fieldName)
	}
	return tfMap, nil
}

func multipartConfigHasPlaceholders(config *MultipartConfigModel) bool {
	if !multipartConfigIsSet(config) {
		return false
	}
	parts, diags := parseMultipartParts(config.Parts)
	if diags.HasError() {
		return false
	}
	for _, part := range parts {
		if hasValue(part.Value) && containsResponsePlaceholder(part.Value.ValueString()) {
			return true
		}
	}
	return false
}

func substituteMultipartPartValues(config *MultipartConfigModel, responseJSON, fieldName string) (*MultipartConfigModel, error) {
	if !multipartConfigIsSet(config) {
		return config, nil
	}

	parts, diags := parseMultipartParts(config.Parts)
	if diags.HasError() {
		return nil, fmt.Errorf("%s: failed to parse multipart parts", fieldName)
	}

	values := make([]attr.Value, 0, len(parts))
	for i, part := range parts {
		value := part.Value.ValueString()
		if hasValue(part.Value) {
			substituted, err := substituteFieldPlaceholders(value, responseJSON, fmt.Sprintf("%s.parts[%d].value", fieldName, i))
			if err != nil {
				return nil, err
			}
			part.Value = types.StringValue(substituted)
		}
		obj, objDiags := types.ObjectValue(multipartPartAttrTypes, map[string]attr.Value{
			"name":         part.Name,
			"value":        part.Value,
			"file_path":    part.FilePath,
			"content_type": part.ContentType,
		})
		if objDiags.HasError() {
			return nil, fmt.Errorf("%s: failed to build multipart part object", fieldName)
		}
		values = append(values, obj)
	}

	partsList, listDiags := types.ListValue(multipartObjectType(), values)
	if listDiags.HasError() {
		return nil, fmt.Errorf("%s: failed to build multipart parts list", fieldName)
	}
	return &MultipartConfigModel{Parts: partsList}, nil
}

func resolveDestroyTemplates(data *CurlResourceModel) (resolvedDestroyTemplates, diag.Diagnostics) {
	var diags diag.Diagnostics
	var resolved resolvedDestroyTemplates

	if data == nil {
		return resolved, diags
	}

	destroyMultipart := data.DestroyRequestMultipart
	payload, payloadDiags := resolveRequestPayload(data.DestroyRequestBody, data.DestroyRequestBodyWo, data.DestroyRequestBodyFile, destroyMultipart)
	diags.Append(payloadDiags...)
	if diags.HasError() {
		return resolved, diags
	}

	resolved.URL = data.DestroyUrl.ValueString()
	resolved.Payload = payload
	resolved.DestroyHeaders = data.DestroyHeaders
	resolved.DestroyHeadersWo = data.DestroyHeadersWo
	resolved.DestroyParams = data.DestroyRequestParameters

	if !destroyFieldsHavePlaceholders(data) {
		return resolved, diags
	}

	responseJSON := priorResponseValue(data.ResponseSensitive, data.Response, data.SensitiveResponse)
	if responseJSON == "" {
		diags.AddError(
			"Destroy Template Error",
			"Destroy configuration contains {response...} placeholders but the stored create response is empty.",
		)
		return resolved, diags
	}

	url, err := substituteFieldPlaceholders(data.DestroyUrl.ValueString(), responseJSON, "destroy_url")
	if err != nil {
		diags.AddError("Destroy Template Error", err.Error())
		return resolved, diags
	}
	resolved.URL = url

	substitutedMultipart, err := substituteMultipartPartValues(data.DestroyRequestMultipart, responseJSON, "destroy_request_multipart")
	if err != nil {
		diags.AddError("Destroy Template Error", err.Error())
		return resolved, diags
	}

	payload, payloadDiags = resolveRequestPayload(data.DestroyRequestBody, data.DestroyRequestBodyWo, data.DestroyRequestBodyFile, substitutedMultipart)
	diags.Append(payloadDiags...)
	if diags.HasError() {
		return resolved, diags
	}
	resolved.Payload = payload

	if len(payload.Body) > 0 && !payload.UsedMultipartBody && !payload.UsedFileBody {
		bodyStr, err := substituteFieldPlaceholders(string(payload.Body), responseJSON, "destroy_request_body")
		if err != nil {
			diags.AddError("Destroy Template Error", err.Error())
			return resolved, diags
		}
		resolved.Payload.Body = []byte(bodyStr)
	}

	headers, err := substituteMapPlaceholders(data.DestroyHeaders, responseJSON, "destroy_headers")
	if err != nil {
		diags.AddError("Destroy Template Error", err.Error())
		return resolved, diags
	}
	resolved.DestroyHeaders = headers

	headersWo, err := substituteMapPlaceholders(data.DestroyHeadersWo, responseJSON, "destroy_headers_wo")
	if err != nil {
		diags.AddError("Destroy Template Error", err.Error())
		return resolved, diags
	}
	resolved.DestroyHeadersWo = headersWo

	params, err := substituteMapPlaceholders(data.DestroyRequestParameters, responseJSON, "destroy_request_parameters")
	if err != nil {
		diags.AddError("Destroy Template Error", err.Error())
		return resolved, diags
	}
	resolved.DestroyParams = params

	return resolved, diags
}

func validateDestroyTemplates(data CurlResourceModel) diag.Diagnostics {
	if data.SkipDestroy.ValueBool() {
		return nil
	}
	_, diags := resolveDestroyTemplates(&data)
	return diags
}
