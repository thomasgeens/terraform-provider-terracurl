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

resource "terracurl_request" "multipart_upload" {
  name   = "multipart-upload"
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
