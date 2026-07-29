---
page_title: "Write-Only Headers and Request Bodies"
subcategory: "Guides"
description: |-
  Use write-only headers and request bodies on terracurl_request resources to send secrets without persisting them in Terraform state.
---

# Write-Only Headers and Request Bodies

Terraform 1.11+ supports **write-only arguments**: values supplied in configuration that are used during apply but are **not stored in plan or state JSON**. TerraCurl exposes this for `terracurl_request` resource headers and request bodies on create, read, and destroy operations.

Use write-only attributes when secrets must not appear in `terraform show`, remote state, or version control–backed state files — even when marked `sensitive`.

## Requirements

- Terraform **1.11 or later**
- `terracurl_request` **resource only** (write-only is not available on data sources, actions, or ephemeral resources)

## Create example

```terraform
resource "terracurl_request" "example" {
  name   = "example"
  url    = "https://api.example.com/resource"
  method = "PUT"

  headers_wo = {
    X-Vault-Token = var.vault_token
  }
  headers_wo_version = 1

  request_body_wo = jsonencode({
    password = var.password
  })
  request_body_wo_version = 1

  response_codes = [200]
  skip_read      = true
}
```

Do **not** set both `headers` and `headers_wo` (or `request_body` and `request_body_wo`) on the same operation. TerraCurl rejects conflicting pairs at plan time.

## Per-operation attributes

| Operation | Headers | Body | Version attributes |
|-----------|---------|------|--------------------|
| Create | `headers_wo` | `request_body_wo` | `headers_wo_version`, `request_body_wo_version` |
| Read | `read_headers_wo` | `read_request_body_wo` | `read_headers_wo_version`, `read_request_body_wo_version` |
| Destroy | `destroy_headers_wo` | `destroy_request_body_wo` | `destroy_headers_wo_version`, `destroy_request_body_wo_version` |

Version attributes are stored in state (not write-only). Increment a `_wo_version` when you change the corresponding write-only value so TerraCurl refreshes the snapshotted credentials on update.

## Read and destroy without config access

Terraform does not pass resource configuration to Read or Delete RPCs. TerraCurl **snapshots** read/destroy write-only values into provider private state during Create (and refreshes the snapshot on Update when a `_wo_version` changes). Private state is not shown in plan output, but it is persisted with the resource.

For rotating destroy credentials without storing them in resource state, consider [provider `default_headers`](default_headers) instead.

## Updating write-only values

TerraCurl does not re-issue the create HTTP call on in-place update. To change a write-only body or header after initial apply:

1. Update the write-only value in configuration
2. Increment the matching `_wo_version` attribute
3. Run `terraform apply`

If the API must be called again with the new payload, trigger replacement through other attribute changes or taint the resource.

## Comparison with other options

| Approach | In state? | In plan? | Destroy refresh |
|----------|-----------|----------|-----------------|
| `headers` / `request_body` | Yes (sensitive still stored) | Hidden if sensitive | From state (can go stale) |
| `headers_wo` / `request_body_wo` | No | No | From private snapshot |
| Provider `default_headers` | No (provider config) | No | Re-evaluated every run |

## Limitations

- Resource-only; use `default_headers` or sensitive attributes on data sources, actions, and ephemeral resources
- Private state snapshot is not zero-persistence storage
- Digest credentials remain stateful today; write-only digest auth is future work
