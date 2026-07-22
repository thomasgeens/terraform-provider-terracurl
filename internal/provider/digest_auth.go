package provider

import (
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/icholy/digest"
)

// DigestAuthModel describes digest authentication credentials from Terraform configuration.
type DigestAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// DigestAuthConfig holds resolved digest authentication credentials for HTTP clients.
type DigestAuthConfig struct {
	Username string
	Password string
}

func digestAuthConfigured(model DigestAuthModel) bool {
	return hasValue(model.Username)
}

func parseDigestAuthConfig(model *DigestAuthModel) *DigestAuthConfig {
	if model == nil || !digestAuthConfigured(*model) {
		return nil
	}
	return &DigestAuthConfig{
		Username: model.Username.ValueString(),
		Password: model.Password.ValueString(),
	}
}

func resolveDigestAuth(provider *DigestAuthConfig, override *DigestAuthModel) *DigestAuthConfig {
	if configured := parseDigestAuthConfig(override); configured != nil {
		return configured
	}
	return provider
}

func wrapTransportWithDigest(base http.RoundTripper, digestAuth *DigestAuthConfig) http.RoundTripper {
	if digestAuth == nil || digestAuth.Username == "" {
		return base
	}
	return &digest.Transport{
		Username:  digestAuth.Username,
		Password:  digestAuth.Password,
		Transport: base,
	}
}

func digestAuthToPrivateMap(model *DigestAuthModel) map[string]string {
	if model == nil || !digestAuthConfigured(*model) {
		return nil
	}
	return map[string]string{
		"username": model.Username.ValueString(),
		"password": model.Password.ValueString(),
	}
}

func digestAuthFromPrivateMap(raw interface{}) *DigestAuthModel {
	if raw == nil {
		return nil
	}
	privateMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	username, _ := privateMap["username"].(string)
	password, _ := privateMap["password"].(string)
	if username == "" {
		return nil
	}
	return &DigestAuthModel{
		Username: types.StringValue(username),
		Password: types.StringValue(password),
	}
}
