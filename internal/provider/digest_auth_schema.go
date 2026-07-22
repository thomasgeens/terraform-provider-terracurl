package provider

import (
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const digestAuthUsernameDescription = "Username for HTTP Digest authentication."
const digestAuthPasswordDescription = "Password for HTTP Digest authentication."

func providerDigestAuthSchema() providerschema.SingleNestedAttribute {
	return providerschema.SingleNestedAttribute{
		MarkdownDescription: "Default HTTP Digest authentication credentials applied to every outbound request when an operation-specific block is not configured.",
		Optional:            true,
		Sensitive:           true,
		Attributes: map[string]providerschema.Attribute{
			"username": providerschema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: digestAuthUsernameDescription,
			},
			"password": providerschema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: digestAuthPasswordDescription,
			},
		},
	}
}

func resourceDigestAuthSchema(markdown string) resourceschema.SingleNestedAttribute {
	return resourceschema.SingleNestedAttribute{
		MarkdownDescription: markdown,
		Optional:            true,
		Sensitive:           true,
		Attributes: map[string]resourceschema.Attribute{
			"username": resourceschema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: digestAuthUsernameDescription,
			},
			"password": resourceschema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: digestAuthPasswordDescription,
			},
		},
	}
}

func dataSourceDigestAuthSchema(markdown string) datasourceschema.SingleNestedAttribute {
	return datasourceschema.SingleNestedAttribute{
		MarkdownDescription: markdown,
		Optional:            true,
		Sensitive:           true,
		Attributes: map[string]datasourceschema.Attribute{
			"username": datasourceschema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: digestAuthUsernameDescription,
			},
			"password": datasourceschema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: digestAuthPasswordDescription,
			},
		},
	}
}

func actionDigestAuthSchema(markdown string) actionschema.SingleNestedAttribute {
	return actionschema.SingleNestedAttribute{
		MarkdownDescription: markdown,
		Optional:            true,
		Attributes: map[string]actionschema.Attribute{
			"username": actionschema.StringAttribute{
				Required:            true,
				MarkdownDescription: digestAuthUsernameDescription,
			},
			"password": actionschema.StringAttribute{
				Required:            true,
				MarkdownDescription: digestAuthPasswordDescription,
			},
		},
	}
}

func ephemeralDigestAuthSchema(markdown string) ephemeralschema.SingleNestedAttribute {
	return ephemeralschema.SingleNestedAttribute{
		MarkdownDescription: markdown,
		Optional:            true,
		Sensitive:           true,
		Attributes: map[string]ephemeralschema.Attribute{
			"username": ephemeralschema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: digestAuthUsernameDescription,
			},
			"password": ephemeralschema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: digestAuthPasswordDescription,
			},
		},
	}
}
