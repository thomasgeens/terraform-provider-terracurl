---
page_title: "HTTP Digest Authentication"
subcategory: "Guides"
description: |-
  Configure provider-level default_digest_auth and per-operation digest_auth overrides for APIs that require HTTP Digest authentication.
---

# HTTP Digest Authentication

Some APIs require [HTTP Digest authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication#digest_authentication) instead of a static `Authorization` header. Digest is a challenge-response protocol: the server returns `401 Unauthorized` with a `WWW-Authenticate: Digest ...` header, and the client signs a follow-up request.

TerraCurl implements Digest at the HTTP transport layer using a digest-aware `RoundTripper`. You do not set `Authorization` manually for Digest auth.

## Provider default

Configure credentials once in the provider block. Provider configuration is re-evaluated on every Terraform run, including destroy.

```terraform
provider "terracurl" {
  default_digest_auth {
    username = var.api_user
    password = var.api_password
  }
}
```

When an operation-specific block is not configured, TerraCurl uses `default_digest_auth` for that request.

## Per-operation overrides

Override credentials on individual operations when create, read, and destroy need different users:

```terraform
resource "terracurl_request" "example" {
  name   = "example"
  url    = "https://api.example.com/resource"
  method = "PUT"

  digest_auth {
    username = "admin"
    password = var.password
  }

  read_digest_auth {
    username = "reader"
    password = var.read_password
  }

  destroy_digest_auth {
    username = "admin"
    password = var.password
  }

  response_codes = [200]
  skip_read      = true

  destroy_url            = "https://api.example.com/resource"
  destroy_method         = "DELETE"
  destroy_response_codes = [200]
}
```

The same nested blocks are available on:

- `terracurl_request` resources (`digest_auth`, `read_digest_auth`, `destroy_digest_auth`)
- `terracurl_request` data sources (`digest_auth`)
- `terracurl_request` actions (`digest_auth`)
- `terracurl_request` ephemeral resources (`digest_auth`, `renew_digest_auth`, `close_digest_auth`)

## Precedence

For each HTTP operation, TerraCurl resolves credentials in this order:

1. Operation-specific block (for example, `destroy_digest_auth`) when `username` is set
2. Provider `default_digest_auth` when configured
3. No Digest authentication (plain HTTP client)

Operation blocks override provider defaults.

## Destroy and stale credentials

During destroy, TerraCurl reads prior resource state for configuration. Provider `default_digest_auth` is **not** stored in resource state—it is re-evaluated every run. Prefer provider-level defaults for destroy when credentials might change between apply and destroy.

If credentials are set only in `destroy_digest_auth`, those values are persisted in state and can go stale. See the [Default Headers for Auth Tokens guide](default_headers) for a similar pattern with short-lived tokens.

## Limitations

- HTTP Basic authentication is not built in; set a static `Authorization: Basic ...` header manually or use provider `default_headers`.
- NTLM, Kerberos, and other enterprise proxy authentication mechanisms remain unsupported. See the [HTTP Proxy Support guide](proxy).
- Digest credentials configured on resources are stored in Terraform state (marked sensitive). Write-only credentials are tracked separately in issue #115.
- Custom `realm` or algorithm tuning is not exposed unless a real API requires it.

## Digest vs default_headers

Use `default_digest_auth` / `digest_auth` for Digest challenge-response authentication. Use `default_headers` for static or bearer-style `Authorization` headers and other custom headers that must stay fresh on destroy.
