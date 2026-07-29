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
