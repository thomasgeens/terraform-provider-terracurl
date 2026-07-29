data "terracurl_request" "binary_download" {
  name           = "binary-download"
  url            = "https://api.example.com/files/report.pdf"
  method         = "GET"
  response_codes = ["200"]
}

# Decode the base64 response and write to disk with the local provider.
resource "local_file" "report" {
  filename = "${path.module}/report.pdf"
  content  = base64decode(data.terracurl_request.binary_download.response_base64)
}
