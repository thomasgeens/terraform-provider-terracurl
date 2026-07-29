data "terracurl_request" "file_body" {
  name           = "file-body"
  url            = "https://api.example.com/data"
  method         = "POST"
  response_codes = ["200"]

  request_body_file = "${path.module}/payload.json"
}

data "terracurl_request" "multipart" {
  name           = "multipart"
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
