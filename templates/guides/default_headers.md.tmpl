---
page_title: "Default Headers for Auth Tokens"
subcategory: "Guides"
description: |-
  Configure provider-level default headers for short-lived authentication tokens that must stay fresh during destroy.
---

# Default Headers for Auth Tokens

Short-lived authentication tokens (OAuth access tokens, GCP ID tokens, session cookies) often expire between `terraform apply` and a later `terraform destroy`. TerraCurl stores resource header values in Terraform state at apply time. During destroy, the provider reads those stored values—not the freshly evaluated configuration—so destroy requests can fail with `401 Unauthorized` if the token has expired.

Provider `default_headers` solves this by applying headers from the provider block on **every** outbound HTTP request. Provider configuration is re-evaluated on each Terraform run, including destroy, so dynamic token expressions stay current.

## When to use default_headers

Use provider `default_headers` when:

- Auth tokens expire quickly (for example, GCP Cloud Run ID tokens, ~1 hour TTL)
- The same auth header is needed on create, read, and destroy
- You reference a data source or variable for the token value

## Example: GCP Cloud Run ID token

```terraform
data "google_service_account_id_token" "sa_gcp" {
  target_service_account = data.google_service_account.sa_id.email
  target_audience        = "https://my-service.run.app/"
}

provider "terracurl" {
  default_headers = {
    "X-Serverless-Authorization" = "Bearer ${data.google_service_account_id_token.sa_gcp.id_token}"
  }
}

resource "terracurl_request" "example" {
  name   = "example"
  url    = "https://my-service.run.app/resource"
  method = "PUT"

  headers = {
    Content-Type = "application/json"
  }

  request_body   = jsonencode({ id = "example" })
  response_codes = [200]
  skip_read      = true

  destroy_url            = "https://my-service.run.app/resource"
  destroy_method         = "DELETE"
  destroy_response_codes = [200]
  destroy_headers = {
    Content-Type = "application/json"
  }
}
```

Move auth headers from `headers` / `destroy_headers` to `default_headers`. Keep non-auth headers (such as `Content-Type`) on the resource if needed.

## Merge behavior

TerraCurl applies headers in this order:

1. Resource-level headers (`headers`, `read_headers`, `destroy_headers`, and so on)
2. Provider `default_headers` (overrides resource headers with the same key)

Provider headers win on key collision. This ensures a fresh provider token replaces a stale token stored in state during destroy.

`Host` (case-insensitive) overrides the HTTP Host header sent on the wire, independent of the URL hostname.

## Scope

`default_headers` applies to all outbound TerraCurl requests:

- `terracurl_request` resources (create, read, destroy)
- `terracurl_request` data sources
- `terracurl_request` actions
- `terracurl_request` ephemeral resources (open, renew, close)

## Limitations

- Tokens set only in resource `headers` or `destroy_headers` are still persisted in state. For destroy-only refresh, move auth to `default_headers` or run `terraform apply` before destroy to update state.
- Write-only resource headers (see issue #115) are a separate follow-up to avoid persisting secrets in state entirely.
- `default_headers` is marked sensitive in the provider schema and will not appear in plan output.

## Workaround without default_headers

If you cannot upgrade yet, run `terraform apply` (with no infrastructure changes) before `terraform destroy`. That updates header values in state with freshly evaluated tokens, then destroy succeeds. This is fragile when `lifecycle { ignore_changes = ... }` blocks header updates.
