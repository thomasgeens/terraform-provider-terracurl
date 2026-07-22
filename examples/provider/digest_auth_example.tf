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
