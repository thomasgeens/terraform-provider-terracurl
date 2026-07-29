package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const writeOnlyPrivateStateKey = "write_only_snapshot"

type writeOnlyPrivateSnapshot struct {
	HeadersCreate  map[string]string `json:"headers_create,omitempty"`
	HeadersRead    map[string]string `json:"headers_read,omitempty"`
	HeadersDestroy map[string]string `json:"headers_destroy,omitempty"`
	BodyCreate     string            `json:"body_create,omitempty"`
	BodyRead       string            `json:"body_read,omitempty"`
	BodyDestroy    string            `json:"body_destroy,omitempty"`
}

type privateStateWriter interface {
	SetKey(context.Context, string, []byte) diag.Diagnostics
}

type privateStateReader interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
}

func mapHasValues(m types.Map) bool {
	if m.IsNull() || m.IsUnknown() {
		return false
	}
	return len(m.Elements()) > 0
}

func mapToStringMap(m types.Map) map[string]string {
	if !mapHasValues(m) {
		return nil
	}
	result := make(map[string]string, len(m.Elements()))
	for k, v := range m.Elements() {
		if strVal, ok := v.(types.String); ok {
			result[k] = strVal.ValueString()
		}
	}
	return result
}

func stringMapToTypesMap(values map[string]string) types.Map {
	if len(values) == 0 {
		return types.MapNull(types.StringType)
	}
	result := make(map[string]types.String, len(values))
	for k, v := range values {
		result[k] = types.StringValue(v)
	}
	tfMap, diags := types.MapValueFrom(context.Background(), types.StringType, result)
	if diags.HasError() {
		return types.MapNull(types.StringType)
	}
	return tfMap
}

func resolveRequestBody(regular, writeOnly types.String) ([]byte, bool) {
	if hasValue(writeOnly) {
		return []byte(writeOnly.ValueString()), true
	}
	if hasValue(regular) {
		return []byte(regular.ValueString()), false
	}
	return nil, false
}

func applyRequestHeadersWithWriteOnly(req *http.Request, headers, headersWo types.Map, meta *ProviderMeta) {
	applyRequestHeaders(req, headers)
	if meta != nil {
		for k, v := range meta.DefaultHeaders() {
			setRequestHeader(req, k, v)
		}
	}
	applyRequestHeaders(req, headersWo)
}

func requestBodyForLog(regular, writeOnly types.String, usedWriteOnly bool) string {
	if usedWriteOnly {
		return "<redacted write-only request body>"
	}
	if hasValue(regular) {
		return regular.ValueString()
	}
	return ""
}

func snapshotFromConfig(config CurlResourceModel) writeOnlyPrivateSnapshot {
	return writeOnlyPrivateSnapshot{
		HeadersCreate:  mapToStringMap(config.HeadersWo),
		HeadersRead:    mapToStringMap(config.ReadHeadersWo),
		HeadersDestroy: mapToStringMap(config.DestroyHeadersWo),
		BodyCreate:     config.RequestBodyWo.ValueString(),
		BodyRead:       config.ReadRequestBodyWo.ValueString(),
		BodyDestroy:    config.DestroyRequestBodyWo.ValueString(),
	}
}

func snapshotWriteOnlyToPrivate(ctx context.Context, config CurlResourceModel, private privateStateWriter) diag.Diagnostics {
	var diags diag.Diagnostics
	if private == nil {
		return diags
	}

	snapshot := snapshotFromConfig(config)
	var raw []byte
	var err error
	if !snapshot.isEmpty() {
		raw, err = json.Marshal(snapshot)
		if err != nil {
			diags.AddError("Write-Only Snapshot Error", fmt.Sprintf("Failed to marshal write-only snapshot: %s", err))
			return diags
		}
	}

	diags.Append(private.SetKey(ctx, writeOnlyPrivateStateKey, raw)...)
	return diags
}

func loadWriteOnlyFromPrivate(ctx context.Context, private privateStateReader, data *CurlResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if private == nil || data == nil {
		return diags
	}

	raw, getDiags := private.GetKey(ctx, writeOnlyPrivateStateKey)
	diags.Append(getDiags...)
	if diags.HasError() || raw == nil {
		return diags
	}

	var snapshot writeOnlyPrivateSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		diags.AddError("Write-Only Snapshot Error", fmt.Sprintf("Failed to unmarshal write-only snapshot: %s", err))
		return diags
	}

	data.HeadersWo = stringMapToTypesMap(snapshot.HeadersCreate)
	data.ReadHeadersWo = stringMapToTypesMap(snapshot.HeadersRead)
	data.DestroyHeadersWo = stringMapToTypesMap(snapshot.HeadersDestroy)

	if snapshot.BodyCreate != "" {
		data.RequestBodyWo = types.StringValue(snapshot.BodyCreate)
	}
	if snapshot.BodyRead != "" {
		data.ReadRequestBodyWo = types.StringValue(snapshot.BodyRead)
	}
	if snapshot.BodyDestroy != "" {
		data.DestroyRequestBodyWo = types.StringValue(snapshot.BodyDestroy)
	}

	return diags
}

func mergeWriteOnlyFromConfig(ctx context.Context, config tfsdk.Config, data *CurlResourceModel) diag.Diagnostics {
	var configModel CurlResourceModel
	diags := config.Get(ctx, &configModel)
	if diags.HasError() || data == nil {
		return diags
	}

	data.HeadersWo = configModel.HeadersWo
	data.RequestBodyWo = configModel.RequestBodyWo
	data.ReadHeadersWo = configModel.ReadHeadersWo
	data.ReadRequestBodyWo = configModel.ReadRequestBodyWo
	data.DestroyHeadersWo = configModel.DestroyHeadersWo
	data.DestroyRequestBodyWo = configModel.DestroyRequestBodyWo

	return diags
}

func (s writeOnlyPrivateSnapshot) isEmpty() bool {
	return len(s.HeadersCreate) == 0 &&
		len(s.HeadersRead) == 0 &&
		len(s.HeadersDestroy) == 0 &&
		s.BodyCreate == "" &&
		s.BodyRead == "" &&
		s.BodyDestroy == ""
}

func writeOnlyVersionsChanged(state, config CurlResourceModel) bool {
	return state.HeadersWoVersion.ValueInt64() != config.HeadersWoVersion.ValueInt64() ||
		state.RequestBodyWoVersion.ValueInt64() != config.RequestBodyWoVersion.ValueInt64() ||
		state.ReadHeadersWoVersion.ValueInt64() != config.ReadHeadersWoVersion.ValueInt64() ||
		state.ReadRequestBodyWoVersion.ValueInt64() != config.ReadRequestBodyWoVersion.ValueInt64() ||
		state.DestroyHeadersWoVersion.ValueInt64() != config.DestroyHeadersWoVersion.ValueInt64() ||
		state.DestroyRequestBodyWoVersion.ValueInt64() != config.DestroyRequestBodyWoVersion.ValueInt64()
}

func applyWriteOnlySnapshot(snapshot writeOnlyPrivateSnapshot, data *CurlResourceModel) {
	if data == nil {
		return
	}
	data.HeadersWo = stringMapToTypesMap(snapshot.HeadersCreate)
	data.ReadHeadersWo = stringMapToTypesMap(snapshot.HeadersRead)
	data.DestroyHeadersWo = stringMapToTypesMap(snapshot.HeadersDestroy)
	if snapshot.BodyCreate != "" {
		data.RequestBodyWo = types.StringValue(snapshot.BodyCreate)
	}
	if snapshot.BodyRead != "" {
		data.ReadRequestBodyWo = types.StringValue(snapshot.BodyRead)
	}
	if snapshot.BodyDestroy != "" {
		data.DestroyRequestBodyWo = types.StringValue(snapshot.BodyDestroy)
	}
}

func nullWriteOnlyAttributes(data *CurlResourceModel) {
	if data == nil {
		return
	}
	data.HeadersWo = types.MapNull(types.StringType)
	data.RequestBodyWo = types.StringNull()
	data.ReadHeadersWo = types.MapNull(types.StringType)
	data.ReadRequestBodyWo = types.StringNull()
	data.DestroyHeadersWo = types.MapNull(types.StringType)
	data.DestroyRequestBodyWo = types.StringNull()
}
