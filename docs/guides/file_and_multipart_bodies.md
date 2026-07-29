---
page_title: "File and Multipart Request Bodies"
subcategory: "Guides"
description: |-
  Load request bodies from disk with request_body_file and build multipart/form-data payloads with request_multipart across TerraCurl surfaces.
---

# File and Multipart Request Bodies

TerraCurl supports two complementary ways to send non-JSON request bodies without embedding large or binary payloads in Terraform configuration or state:

| Mechanism | Use case | Stored in state |
|-----------|----------|-----------------|
| `request_body_file` | Raw binary or text body (for example a ZIP deploy) | File path only |
| `request_multipart` | `multipart/form-data` with text fields and file parts | Part metadata and paths only |

These attributes are available on **resources**, **data sources**, **actions**, and **ephemeral resources**. On `terracurl_request` resources, matching `read_*` and `destroy_*` variants exist for read and destroy operations.

## Mutual exclusion

For each HTTP operation, configure **one** body source:

- `request_body` or `request_body_wo` (resources only for `_wo`)
- `request_body_file`
- `request_multipart`

The provider validates this at plan time.

## File-based request bodies

Use `request_body_file` when the API expects a raw payload such as a ZIP archive, protobuf message, or plain-text document.

```terraform
resource "terracurl_request" "kudu_deploy" {
  name   = "kudu-deploy"
  url    = "https://myapp.scm.azurewebsites.net/api/zipdeploy?isAsync=true"
  method = "PUT"

  request_body_file = "${path.module}/app.zip"

  headers = {
    Content-Type = "application/zip"
  }

  response_codes = [200, 202]
  skip_read      = true
  skip_destroy   = true
}
```

Behavior:

- The provider reads the file with `os.ReadFile` at HTTP call time.
- Only the **path** is stored in Terraform state, not the file bytes (same pattern as `cert_file`).
- Changing the path triggers replace via `RequiresReplace`.
- Content changes at the same path do **not** automatically trigger replace. Taint or change the path when the on-disk file changes.

Set `Content-Type` in `headers` when the API requires a specific media type.

## Multipart form bodies

Use `request_multipart` for APIs that accept `multipart/form-data` uploads (similar to `curl -F name=John -F photo=@john.jpg`).

```terraform
resource "terracurl_request" "upload" {
  name   = "upload"
  url    = "https://api.example.com/upload"
  method = "POST"

  request_multipart = {
    parts = [
      { name = "name", value = "John" },
      { name = "photo", file_path = "${path.module}/john.jpg", content_type = "image/jpeg" },
    ]
  }

  response_codes = [200]
  skip_read      = true
  skip_destroy   = true
}
```

Each part requires:

| Attribute | Required | Description |
|-----------|----------|-------------|
| `name` | Yes | Form field name |
| `value` | One of `value` or `file_path` | Text field value |
| `file_path` | One of `value` or `file_path` | Path to a file on disk |
| `content_type` | No | Defaults to `text/plain` for value parts and `application/octet-stream` for file parts |

The provider builds the body with Go's `mime/multipart` package and sets `Content-Type: multipart/form-data; boundary=...` on the outbound request. If you also set `Content-Type` in `headers`, the provider-generated multipart header **wins** for that operation.

## Data source example

```terraform
data "terracurl_request" "upload_status" {
  name           = "upload-status"
  url            = "https://api.example.com/upload"
  method         = "POST"
  response_codes = ["200"]

  request_multipart = {
    parts = [
      { name = "name", value = "John" },
      { name = "photo", file_path = "${path.module}/john.jpg" },
    ]
  }
}
```

## Action example

```terraform
action "terracurl_request" "notify" {
  url    = "https://api.example.com/notify"
  method = "POST"

  request_body_file = "${path.module}/payload.bin"

  headers = {
    Content-Type = "application/octet-stream"
  }

  response_codes = [200, 204]
}
```

## Ephemeral resource example

Ephemeral resources support `request_body_file` / `request_multipart` on open, plus `renew_*` and `close_*` variants. Renew and close file paths and multipart metadata are snapshotted to private state at open and re-read from disk at renew/close time.

```terraform
ephemeral "terracurl_request" "session" {
  name           = "session"
  url            = "https://api.example.com/open"
  method         = "POST"
  response_codes = ["201"]

  skip_renew = false
  renew_url  = "https://api.example.com/renew"
  renew_method = "PUT"
  renew_request_body_file = "${path.module}/renew-payload.json"
  renew_response_codes = ["200"]

  skip_close = false
  close_url  = "https://api.example.com/close"
  close_method = "DELETE"
  close_request_multipart = {
    parts = [
      { name = "token", value = "placeholder" },
    ]
  }
  close_response_codes = ["204"]
}
```

## Resource read and destroy variants

On `terracurl_request` resources:

- `read_request_body_file` / `read_request_multipart` for read calls
- `destroy_request_body_file` / `destroy_request_multipart` for destroy calls

Write-only string bodies (`*_request_body_wo`) remain available for inline secrets. File and multipart attributes are already path/metadata-only in state, so write-only variants are not provided for them.

## Destroy templating with multipart

`{response.<path>}` placeholders in `destroy_request_multipart` part `value` fields are substituted at destroy time from the stored create response. File parts (`file_path`) are not templated.

See the [Destroy Response Templating guide](destroy_templating) for placeholder syntax.

## Binary response bodies (data source)

To download binary files without UTF-8 corruption, use the data source computed attributes `response_base64` and `sensitive_response_base64`. These encode the raw response bytes with RFC 4648 standard base64. Decode in Terraform with `base64decode()` (for example when writing to disk with `local_file`).

See [`examples/data-sources/binary_response_example`](../../examples/data-sources/binary_response_example/data-source.tf).

## Limitations

- File contents are loaded fully into memory at request time (same as embedding bytes in configuration).
- The provider does not sniff MIME types from file extensions; set `content_type` explicitly when needed.
- Write-only multipart parts are not supported (multipart configuration is already metadata-only in state).

## Related guides

- [Write-Only Headers and Request Bodies](write_only) — inline secret bodies without persisting values in state
- [Destroy Response Templating](destroy_templating) — inject create response values into destroy configuration
