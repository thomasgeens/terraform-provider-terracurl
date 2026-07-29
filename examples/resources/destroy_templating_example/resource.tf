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
