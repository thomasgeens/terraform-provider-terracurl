---
page_title: "Provider: TerraCurl"
description: |-
  TerraCurl is an open-source Terraform provider that enables declarative, configurable HTTP API interactions directly from Terraform. It allows infrastructure and platform teams to integrate with most REST or HTTP-based services including those without native Terraform providers, using standard Terraform workflows.

  With TerraCurl, you can define create, read, update, and delete operations as HTTP requests, complete with support for custom headers, authentication, TLS configuration, retries, and response parsing. This makes it possible to manage third-party APIs, internal services, and bespoke platforms as first-class Terraform resources, without resorting to brittle null_resource hacks or external scripts.

  TerraCurl is designed for reliability and correctness, supporting state reconciliation, drift detection, idempotency, and lifecycle control. It can trigger resource recreation when remote state diverges from expected responses, handle ephemeral resources, and work with multipart and form-based APIs. By bringing arbitrary HTTP endpoints under Terraform’s declarative model, TerraCurl bridges the gap between “infrastructure as code” and “API as code”, enabling consistent automation, auditability, and repeatability across the entire platform stack.
---

# TERRACURL Provider

TerraCurl is an open-source Terraform provider that enables declarative, configurable HTTP API interactions directly from Terraform. It allows infrastructure and platform teams to integrate with most REST or HTTP-based services, including those without native Terraform providers, using standard Terraform workflows.

With TerraCurl, you can define create, read, update, and delete operations as HTTP requests, complete with support for custom headers, authentication, TLS configuration, retries, and response parsing. This makes it possible to manage third-party APIs, internal services, and bespoke platforms as first-class Terraform resources, without resorting to brittle null_resource hacks or external scripts.

TerraCurl is designed for reliability and correctness, supporting state reconciliation, drift detection, idempotency, and lifecycle control. It can trigger resource recreation when remote state diverges from expected responses, handle ephemeral resources, and work with multipart and form-based APIs. By bringing arbitrary HTTP endpoints under Terraform’s declarative model, TerraCurl bridges the gap between “infrastructure as code” and “API as code”, enabling consistent automation, auditability, and repeatability across the entire platform stack.

Use the navigation to the left to read about the available resources.

## Example Usage

As of Terraform 1.8 and later, providers can implement functions that you can call from the Terraform configuration.

Define the provider as a `required_provider` to use its functions

```terraform
terraform {
  required_providers {
    terracurl = {
      source  = "devops-rob/terracurl"
      version = "2.0.0"
    }
  }
}

provider "terracurl" {}
```

## HTTP Proxy Support

TerraCurl supports outbound HTTP and HTTPS proxies through standard environment variables (`HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`) and optional provider attributes (`http_proxy`, `https_proxy`, `no_proxy`).

See the [HTTP Proxy Support guide](guides/proxy) for configuration examples, precedence rules, and limitations.

```terraform
# Proxy can be configured with provider attributes:
#
# provider "terracurl" {
#   http_proxy  = "http://proxy.example.com:8080"
#   https_proxy = "http://proxy.example.com:8080"
#   no_proxy    = "localhost,127.0.0.1,.internal.example.com"
# }
#
# Alternatively, configure the environment where Terraform runs:
#
# export HTTP_PROXY="http://proxy.example.com:8080"
# export HTTPS_PROXY="http://proxy.example.com:8080"
# export NO_PROXY="localhost,127.0.0.1,.internal.example.com"

provider "terracurl" {
  http_proxy  = "http://proxy.example.com:8080"
  https_proxy = "http://proxy.example.com:8080"
  no_proxy    = "localhost,127.0.0.1,.internal.example.com"
}
```

## Default Headers for Auth Tokens

TerraCurl supports provider-level `default_headers` for short-lived authentication tokens that must be refreshed on every Terraform run, including destroy.

See the [Default Headers for Auth Tokens guide](guides/default_headers) for configuration examples, merge behavior, and migration from resource-level auth headers.

```terraform
# Provider default_headers can be configured in the provider block:
#
# provider "terracurl" {
#   default_headers = {
#     Authorization = "Bearer ${var.api_token}"
#   }
# }
#
# Use default_headers for short-lived auth tokens that must be refreshed on
# every Terraform run, including destroy. Provider headers override resource
# headers with the same key.

provider "terracurl" {
  default_headers = {
    Authorization = "Bearer example-token"
  }
}

resource "terracurl_request" "example" {
  name   = "example"
  url    = "https://httpbin.org/put"
  method = "PUT"

  headers = {
    Content-Type = "application/json"
  }

  request_body   = jsonencode({ id = "example" })
  response_codes = [200]
  skip_read      = true

  destroy_url            = "https://httpbin.org/delete"
  destroy_method         = "DELETE"
  destroy_response_codes = [200]
  destroy_headers = {
    Content-Type = "application/json"
  }
}
```

## HTTP Digest Authentication

TerraCurl supports HTTP Digest authentication with provider-level `default_digest_auth` and per-operation overrides (`digest_auth`, `read_digest_auth`, `destroy_digest_auth`, and related blocks on data sources, actions, and ephemeral resources).

See the [HTTP Digest Authentication guide](guides/digest_auth) for configuration examples, precedence rules, and destroy guidance.

```terraform
# Provider default_digest_auth can be configured in the provider block:
#
# provider "terracurl" {
#   default_digest_auth {
#     username = var.api_user
#     password = var.api_password
#   }
# }
#
# Use default_digest_auth when the same Digest credentials apply to every
# operation. Override per operation with digest_auth, read_digest_auth, or
# destroy_digest_auth on the resource.

provider "terracurl" {
  default_digest_auth {
    username = "admin"
    password = "example-password"
  }
}

resource "terracurl_request" "example" {
  name   = "example"
  url    = "https://httpbin.org/put"
  method = "PUT"

  digest_auth {
    username = "writer"
    password = "writer-password"
  }

  headers = {
    Content-Type = "application/json"
  }

  request_body   = jsonencode({ id = "example" })
  response_codes = [200]
  skip_read      = true

  destroy_url            = "https://httpbin.org/delete"
  destroy_method         = "DELETE"
  destroy_response_codes = [200]
}
```

## Write-Only Headers and Request Bodies

TerraCurl supports write-only `headers_wo` and `request_body_wo` attributes on `terracurl_request` resources (Terraform 1.11+). Values are sent on HTTP calls but are not stored in Terraform state.

See the [Write-Only Headers and Request Bodies guide](guides/write_only) for version attributes, read/destroy private-state behavior, and comparison with `default_headers`.

```terraform
# Write-only headers and bodies require Terraform 1.11+.
#
# Values are used for HTTP calls but not stored in Terraform state.
# Increment *_wo_version when changing a write-only value.

resource "terracurl_request" "example" {
  name   = "example"
  url    = "https://httpbin.org/put"
  method = "PUT"

  headers_wo = {
    Authorization = "Bearer ${var.api_token}"
  }
  headers_wo_version = 1

  request_body_wo = jsonencode({
    password = var.password
  })
  request_body_wo_version = 1

  response_codes = [200]
  skip_read      = true

  destroy_url            = "https://httpbin.org/delete"
  destroy_method         = "DELETE"
  destroy_response_codes = [200]

  destroy_headers_wo = {
    Authorization = "Bearer ${var.api_token}"
  }
  destroy_headers_wo_version = 1
}
```

## Destroy Response Templating

TerraCurl resolves `{response.<path>}` placeholders in destroy URL, body, headers, and query parameters from the stored create response at delete time. This supports APIs that return a server-generated ID needed for delete without Terraform self-references.

See the [Destroy Response Templating guide](guides/destroy_templating) for syntax, nested paths, body-based delete, and create-time validation.

```terraform
resource "terracurl_request" "object" {
  name           = "example-object"
  url            = "https://api.example.com/objects"
  method         = "POST"
  request_body   = jsonencode({ name = "example" })
  response_codes = [201]
  skip_read      = true
  skip_destroy   = false

  destroy_url            = "https://api.example.com/objects/{response.id}"
  destroy_method         = "DELETE"
  destroy_response_codes = [200, 204]
}

resource "terracurl_request" "nested_object" {
  name           = "nested-object"
  url            = "https://api.example.com/objects"
  method         = "POST"
  request_body   = jsonencode({ name = "example" })
  response_codes = [201]
  skip_read      = true
  skip_destroy   = false

  destroy_url            = "https://api.example.com/objects/{response.data.object_id}"
  destroy_method         = "DELETE"
  destroy_response_codes = [200]
}

resource "terracurl_request" "body_destroy" {
  name           = "body-destroy"
  url            = "https://api.example.com/users"
  method         = "POST"
  request_body   = jsonencode({ name = "john" })
  response_codes = [201]
  skip_read      = true
  skip_destroy   = false

  destroy_url            = "https://api.example.com/deactivate"
  destroy_method         = "POST"
  destroy_request_body   = jsonencode({ user_id = "{response.data.user.uuid}" })
  destroy_response_codes = [200]
}
```

## Limitations
