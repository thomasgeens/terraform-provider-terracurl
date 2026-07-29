package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type MultipartPartModel struct {
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	FilePath    types.String `tfsdk:"file_path"`
	ContentType types.String `tfsdk:"content_type"`
}

type MultipartConfigModel struct {
	Parts types.List `tfsdk:"parts"`
}

var multipartPartAttrTypes = map[string]attr.Type{
	"name":         types.StringType,
	"value":        types.StringType,
	"file_path":    types.StringType,
	"content_type": types.StringType,
}

type multipartPartJSON struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type multipartConfigJSON struct {
	Parts []multipartPartJSON `json:"parts,omitempty"`
}

func multipartConfigIsSet(config *MultipartConfigModel) bool {
	if config == nil || config.Parts.IsNull() || config.Parts.IsUnknown() {
		return false
	}
	return len(config.Parts.Elements()) > 0
}

func parseMultipartParts(parts types.List) ([]MultipartPartModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if parts.IsNull() || parts.IsUnknown() {
		return nil, diags
	}

	result := make([]MultipartPartModel, 0, len(parts.Elements()))
	for i, element := range parts.Elements() {
		obj, ok := element.(types.Object)
		if !ok {
			diags.AddError("Multipart Configuration Error", fmt.Sprintf("parts[%d] is not an object", i))
			return nil, diags
		}

		var part MultipartPartModel
		diags.Append(obj.As(ctxBackground, &part, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		result = append(result, part)
	}
	return result, diags
}

var ctxBackground = context.Background()

func validateMultipartParts(parts []MultipartPartModel, fieldPrefix string) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(parts) == 0 {
		return diags
	}

	for i, part := range parts {
		if part.Name.IsNull() || part.Name.ValueString() == "" {
			diags.AddError("Multipart Configuration Error", fmt.Sprintf("%sparts[%d].name is required", fieldPrefix, i))
			continue
		}

		hasTextValue := hasValue(part.Value)
		hasFile := hasValue(part.FilePath)
		switch {
		case hasTextValue && hasFile:
			diags.AddError("Multipart Configuration Error", fmt.Sprintf("%sparts[%d]: specify either value or file_path, not both", fieldPrefix, i))
		case !hasTextValue && !hasFile:
			diags.AddError("Multipart Configuration Error", fmt.Sprintf("%sparts[%d]: either value or file_path is required", fieldPrefix, i))
		}
	}
	return diags
}

func multipartConfigFromJSON(raw []byte) (*MultipartConfigModel, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var payload multipartConfigJSON
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	if len(payload.Parts) == 0 {
		return nil, nil
	}

	values := make([]attr.Value, 0, len(payload.Parts))
	for _, part := range payload.Parts {
		obj, diags := types.ObjectValue(multipartPartAttrTypes, map[string]attr.Value{
			"name":         types.StringValue(part.Name),
			"value":        types.StringValue(part.Value),
			"file_path":    types.StringValue(part.FilePath),
			"content_type": types.StringValue(part.ContentType),
		})
		if diags.HasError() {
			return nil, fmt.Errorf("failed to build multipart part object")
		}
		values = append(values, obj)
	}

	parts, diags := types.ListValue(types.ObjectType{AttrTypes: multipartPartAttrTypes}, values)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to build multipart parts list")
	}
	return &MultipartConfigModel{Parts: parts}, nil
}

func multipartConfigToJSON(config *MultipartConfigModel) ([]byte, error) {
	if !multipartConfigIsSet(config) {
		return nil, nil
	}

	parts, diags := parseMultipartParts(config.Parts)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse multipart parts")
	}

	payload := multipartConfigJSON{Parts: make([]multipartPartJSON, 0, len(parts))}
	for _, part := range parts {
		payload.Parts = append(payload.Parts, multipartPartJSON{
			Name:        part.Name.ValueString(),
			Value:       part.Value.ValueString(),
			FilePath:    part.FilePath.ValueString(),
			ContentType: part.ContentType.ValueString(),
		})
	}
	return json.Marshal(payload)
}

func multipartConfigToPrivateString(config *MultipartConfigModel) (string, error) {
	raw, err := multipartConfigToJSON(config)
	if err != nil {
		return "", err
	}
	if raw == nil {
		return "", nil
	}
	return string(raw), nil
}

func multipartConfigFromPrivateString(value string) (*MultipartConfigModel, error) {
	if value == "" {
		return nil, nil
	}
	return multipartConfigFromJSON([]byte(value))
}
