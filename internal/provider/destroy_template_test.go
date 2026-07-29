package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestContainsResponsePlaceholder(t *testing.T) {
	if containsResponsePlaceholder("https://example.com/{response.id}") != true {
		t.Fatal("expected placeholder detection")
	}
	if containsResponsePlaceholder("https://example.com/static") != false {
		t.Fatal("expected no placeholder")
	}
}

func TestSubstituteResponsePlaceholders_Passthrough(t *testing.T) {
	got, err := substituteResponsePlaceholders("https://example.com/destroy", `{"id":"abc"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com/destroy" {
		t.Fatalf("got %q", got)
	}
}

func TestSubstituteResponsePlaceholders_TopLevel(t *testing.T) {
	got, err := substituteResponsePlaceholders(
		"https://example.com/objects/{response.id}",
		`{"id":"uuid-123"}`,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com/objects/uuid-123" {
		t.Fatalf("got %q", got)
	}
}

func TestSubstituteResponsePlaceholders_Nested(t *testing.T) {
	response := `{"data":{"object_id":"nested-456"}}`
	got, err := substituteResponsePlaceholders(
		"https://example.com/objects/{response.data.object_id}",
		response,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com/objects/nested-456" {
		t.Fatalf("got %q", got)
	}
}

func TestSubstituteResponsePlaceholders_ArrayIndex(t *testing.T) {
	response := `{"items":[{"id":"first-id"}]}`
	got, err := substituteResponsePlaceholders(
		"https://example.com/{response.items.0.id}",
		response,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com/first-id" {
		t.Fatalf("got %q", got)
	}
}

func TestSubstituteResponsePlaceholders_Multiple(t *testing.T) {
	response := `{"tenant_id":"t1","user":{"id":"u1"}}`
	got, err := substituteResponsePlaceholders(
		"https://example.com/tenants/{response.tenant_id}/users/{response.user.id}",
		response,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com/tenants/t1/users/u1" {
		t.Fatalf("got %q", got)
	}
}

func TestSubstituteResponsePlaceholders_EscapeDoubleBraces(t *testing.T) {
	got, err := substituteResponsePlaceholders(
		"literal {{response.id}} value",
		`{"id":"ignored"}`,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "literal {response.id} value" {
		t.Fatalf("got %q", got)
	}
}

func TestSubstituteResponsePlaceholders_MissingPath(t *testing.T) {
	_, err := substituteResponsePlaceholders(
		"https://example.com/{response.missing}",
		`{"id":"abc"}`,
	)
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestSubstituteResponsePlaceholders_NonJSONResponse(t *testing.T) {
	_, err := substituteResponsePlaceholders(
		"https://example.com/{response.id}",
		"not json",
	)
	if err == nil {
		t.Fatal("expected error for non-JSON response")
	}
}

func TestExtractJSONPath_NumericAndBool(t *testing.T) {
	response := `{"count":42,"active":true}`

	count, err := extractJSONPath(response, "count")
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != "42" {
		t.Fatalf("count got %q", count)
	}

	active, err := extractJSONPath(response, "active")
	if err != nil {
		t.Fatalf("active: %v", err)
	}
	if active != "true" {
		t.Fatalf("active got %q", active)
	}
}

func TestResolveDestroyTemplates_NoPlaceholders(t *testing.T) {
	data := &CurlResourceModel{
		DestroyUrl:         types.StringValue("https://example.com/destroy"),
		DestroyRequestBody: types.StringValue(`{"static":true}`),
	}

	resolved, diags := resolveDestroyTemplates(data)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if resolved.URL != "https://example.com/destroy" {
		t.Fatalf("url got %q", resolved.URL)
	}
	if string(resolved.Payload.Body) != `{"static":true}` {
		t.Fatalf("body got %q", string(resolved.Payload.Body))
	}
}

func TestResolveDestroyTemplates_WithNestedPlaceholder(t *testing.T) {
	data := &CurlResourceModel{
		Response:           types.StringValue(`{"data":{"object_id":"nested-456"}}`),
		DestroyUrl:         types.StringValue("https://example.com/objects/{response.data.object_id}"),
		DestroyRequestBody: types.StringValue(`{"id":"{response.data.object_id}"}`),
	}

	resolved, diags := resolveDestroyTemplates(data)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if resolved.URL != "https://example.com/objects/nested-456" {
		t.Fatalf("url got %q", resolved.URL)
	}
	if string(resolved.Payload.Body) != `{"id":"nested-456"}` {
		t.Fatalf("body got %q", string(resolved.Payload.Body))
	}
}

func TestValidateDestroyTemplates_SkipDestroy(t *testing.T) {
	data := CurlResourceModel{
		SkipDestroy: types.BoolValue(true),
		DestroyUrl:  types.StringValue("https://example.com/{response.id}"),
	}
	diags := validateDestroyTemplates(data)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
}

func TestResolveDestroyTemplates_MultipartValuePlaceholder(t *testing.T) {
	data := &CurlResourceModel{
		Response:    types.StringValue(`{"id":"abc-123"}`),
		DestroyUrl:  types.StringValue("https://example.com/destroy"),
		DestroyMethod: types.StringValue("POST"),
		DestroyRequestMultipart: multipartConfigFromParts(t, []map[string]string{
			{"name": "id", "value": "{response.id}"},
		}),
	}

	resolved, diags := resolveDestroyTemplates(data)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if !strings.Contains(string(resolved.Payload.Body), "abc-123") {
		t.Fatalf("multipart body got %q", resolved.Payload.Body)
	}
}
