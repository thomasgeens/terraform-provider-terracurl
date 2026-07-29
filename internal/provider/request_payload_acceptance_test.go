package provider

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jarcoal/httpmock"
)

func TestAccResourceRequestBodyFile(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	bodyFile := filepath.Join(t.TempDir(), "app.zip")
	wantBody := []byte("PK\x03\x04test-zip")
	if err := os.WriteFile(bodyFile, wantBody, 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}

	var gotBody []byte
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"PUT",
		"https://example.com/deploy",
		func(req *http.Request) (*http.Response, error) {
			gotBody, _ = io.ReadAll(req.Body)
			return httpmock.NewStringResponse(200, `{"ok":true}`), nil
		},
	)

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRequestBodyFile(rName, bodyFile),
				Check: resource.ComposeTestCheckFunc(
					func(*terraform.State) error {
						if string(gotBody) != string(wantBody) {
							return fmt.Errorf("request body got %q want %q", gotBody, wantBody)
						}
						return nil
					},
					resource.TestCheckResourceAttr("terracurl_request.file_body", "request_body_file", bodyFile),
				),
			},
		},
	})
}

func testAccResourceRequestBodyFile(name, bodyFile string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "file_body" {
  name           = "%s"
  url            = "https://example.com/deploy"
  method         = "PUT"
  response_codes = ["200"]

  request_body_file = "%s"

  headers = {
    Content-Type = "application/zip"
  }

  skip_read    = true
  skip_destroy = true
}
`, name, bodyFile)
}

func TestAccResourceRequestMultipart(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	attachmentFile := filepath.Join(t.TempDir(), "photo.jpg")
	if err := os.WriteFile(attachmentFile, []byte("jpeg-bytes"), 0o600); err != nil {
		t.Fatalf("write attachment file: %v", err)
	}

	var gotContentType string
	var gotBody string
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/upload",
		func(req *http.Request) (*http.Response, error) {
			gotContentType = req.Header.Get("Content-Type")
			bodyBytes, _ := io.ReadAll(req.Body)
			gotBody = string(bodyBytes)
			return httpmock.NewStringResponse(200, `{"uploaded":true}`), nil
		},
	)

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRequestMultipart(rName, attachmentFile),
				Check: resource.ComposeTestCheckFunc(
					func(*terraform.State) error {
						if !strings.HasPrefix(gotContentType, "multipart/form-data; boundary=") {
							return fmt.Errorf("content type got %q", gotContentType)
						}
						if !strings.Contains(gotBody, `name="name"`) || !strings.Contains(gotBody, "John") {
							return fmt.Errorf("multipart body missing text field: %q", gotBody)
						}
						if !strings.Contains(gotBody, `filename="photo.jpg"`) || !strings.Contains(gotBody, "jpeg-bytes") {
							return fmt.Errorf("multipart body missing file field: %q", gotBody)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccResourceRequestMultipart(name, attachmentFile string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "multipart" {
  name           = "%s"
  url            = "https://example.com/upload"
  method         = "POST"
  response_codes = ["200"]

  request_multipart = {
    parts = [
      { name = "name", value = "John" },
      { name = "photo", file_path = "%s", content_type = "image/jpeg" },
    ]
  }

  skip_read    = true
  skip_destroy = true
}
`, name, attachmentFile)
}

func TestAccDataSourceRequestBodyFile(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	bodyFile := filepath.Join(t.TempDir(), "payload.json")
	wantBody := []byte(`{"from":"file"}`)
	if err := os.WriteFile(bodyFile, wantBody, 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}

	var gotBody []byte
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/data",
		func(req *http.Request) (*http.Response, error) {
			gotBody, _ = io.ReadAll(req.Body)
			return httpmock.NewStringResponse(200, `{"ok":true}`), nil
		},
	)

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRequestBodyFile(rName, bodyFile),
				Check: resource.ComposeTestCheckFunc(
					func(*terraform.State) error {
						if string(gotBody) != string(wantBody) {
							return fmt.Errorf("request body got %q want %q", gotBody, wantBody)
						}
						return nil
					},
					resource.TestCheckResourceAttr("data.terracurl_request.file_body", "request_body_file", bodyFile),
				),
			},
		},
	})
}

func testAccDataSourceRequestBodyFile(name, bodyFile string) string {
	return fmt.Sprintf(`
data "terracurl_request" "file_body" {
  method         = "POST"
  name           = "%s"
  url            = "https://example.com/data"
  response_codes = ["200"]

  request_body_file = "%s"
}
`, name, bodyFile)
}

func TestAccEphemeralRenewClosePayloadFiles(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
	skipIfTerraformIsLegacy(t)

	renewBodyFile := filepath.Join(t.TempDir(), "renew.txt")
	closeBodyFile := filepath.Join(t.TempDir(), "close.txt")
	if err := os.WriteFile(renewBodyFile, []byte("renew-from-file"), 0o600); err != nil {
		t.Fatalf("write renew body file: %v", err)
	}
	if err := os.WriteFile(closeBodyFile, []byte("close-from-file"), 0o600); err != nil {
		t.Fatalf("write close body file: %v", err)
	}

	var renewBody, closeBody string
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://example.com/open", httpmock.NewStringResponder(201, `token-123`))
	httpmock.RegisterResponder("POST", "https://example.com/create", httpmock.NewStringResponder(200, ""))
	httpmock.RegisterResponder(
		"PUT",
		"https://example.com/renew",
		func(req *http.Request) (*http.Response, error) {
			bodyBytes, _ := io.ReadAll(req.Body)
			renewBody = string(bodyBytes)
			return httpmock.NewStringResponse(200, `token-123`), nil
		},
	)
	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/close",
		func(req *http.Request) (*http.Response, error) {
			bodyBytes, _ := io.ReadAll(req.Body)
			closeBody = string(bodyBytes)
			return httpmock.NewStringResponse(204, ""), nil
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
		Steps: []resource.TestStep{
			{
				Config: testAccEphemeralRenewClosePayloadFiles(renewBodyFile, closeBodyFile),
				Check: resource.ComposeTestCheckFunc(
					func(*terraform.State) error {
						if renewBody != "renew-from-file" {
							return fmt.Errorf("renew body got %q want %q", renewBody, "renew-from-file")
						}
						if closeBody != "close-from-file" {
							return fmt.Errorf("close body got %q want %q", closeBody, "close-from-file")
						}
						return nil
					},
					testMockEndpointRegister("PUT https://example.com/renew"),
					testMockEndpointRegister("DELETE https://example.com/close"),
				),
			},
		},
	})
}

func testAccEphemeralRenewClosePayloadFiles(renewBodyFile, closeBodyFile string) string {
	return fmt.Sprintf(`
ephemeral "terracurl_request" "ephems" {
  method         = "POST"
  name           = "test"
  response_codes = ["201"]
  url            = "https://example.com/open"

  skip_renew           = false
  renew_interval       = "-10"
  renew_url            = "https://example.com/renew"
  renew_response_codes = ["200"]
  renew_method         = "PUT"
  renew_request_body_file = "%s"
  renew_timeout        = 20

  skip_close           = false
  close_url            = "https://example.com/close"
  close_response_codes = ["204"]
  close_method         = "DELETE"
  close_request_body_file = "%s"
  close_timeout        = 20
}

resource "terracurl_request" "test" {
  name           = "leader"
  url            = "https://example.com/create"
  response_codes = ["200"]
  method         = "POST"
  skip_destroy   = true
  skip_read      = true
}

provider "echo" {
  data = ephemeral.terracurl_request.ephems
}

resource "echo" "test" {}
`, renewBodyFile, closeBodyFile)
}
