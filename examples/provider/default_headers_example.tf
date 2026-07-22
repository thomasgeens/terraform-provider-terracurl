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
