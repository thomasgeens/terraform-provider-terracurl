ephemeral "terracurl_request" "session" {
  name           = "session"
  url            = "https://api.example.com/open"
  method         = "POST"
  response_codes = ["201"]

  request_body_file = "${path.module}/open-payload.json"

  skip_renew              = false
  renew_url               = "https://api.example.com/renew"
  renew_method            = "PUT"
  renew_request_body_file = "${path.module}/renew-payload.json"
  renew_response_codes    = ["200"]

  skip_close   = false
  close_url    = "https://api.example.com/close"
  close_method = "DELETE"
  close_request_multipart = {
    parts = [
      { name = "reason", value = "done" },
    ]
  }
  close_response_codes = ["204"]
}
