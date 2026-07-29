action "terracurl_request" "multipart_notify" {
  url    = "https://api.example.com/notify"
  method = "POST"

  request_multipart = {
    parts = [
      { name = "message", value = "hello" },
      { name = "attachment", file_path = "${path.module}/report.txt" },
    ]
  }

  response_codes = [200, 204]
}
