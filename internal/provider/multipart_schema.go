package provider

import (
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	requestBodyFileDescription          = "Path to a file on local disk to use as the request body. File bytes are read at request time and are not stored in Terraform state."
	requestMultipartDescription         = "Multipart form request body. The provider sets `Content-Type` with a generated boundary."
	multipartPartNameDescription        = "Form field name."
	multipartPartValueDescription       = "Text form field value."
	multipartPartFilePathDescription    = "Path to a file on local disk for this form field."
	multipartPartContentTypeDescription = "Optional Content-Type for this part. Defaults to `text/plain` for value parts and `application/octet-stream` for file parts."
)

func multipartPartAttributes() map[string]resourceschema.Attribute {
	return map[string]resourceschema.Attribute{
		"name": resourceschema.StringAttribute{
			Required:            true,
			MarkdownDescription: multipartPartNameDescription,
		},
		"value": resourceschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartValueDescription,
		},
		"file_path": resourceschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartFilePathDescription,
		},
		"content_type": resourceschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartContentTypeDescription,
		},
	}
}

func resourceMultipartSchema(markdown string) resourceschema.SingleNestedAttribute {
	return resourceschema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: markdown,
		Attributes: map[string]resourceschema.Attribute{
			"parts": resourceschema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "Multipart form parts.",
				NestedObject: resourceschema.NestedAttributeObject{
					Attributes: multipartPartAttributes(),
				},
			},
		},
	}
}

func resourceRequestBodyFileSchema(markdown string) resourceschema.StringAttribute {
	return resourceschema.StringAttribute{
		Optional:            true,
		MarkdownDescription: markdown,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func dataSourceMultipartPartAttributes() map[string]datasourceschema.Attribute {
	return map[string]datasourceschema.Attribute{
		"name": datasourceschema.StringAttribute{
			Required:            true,
			MarkdownDescription: multipartPartNameDescription,
		},
		"value": datasourceschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartValueDescription,
		},
		"file_path": datasourceschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartFilePathDescription,
		},
		"content_type": datasourceschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartContentTypeDescription,
		},
	}
}

func dataSourceMultipartSchema(markdown string) datasourceschema.SingleNestedAttribute {
	return datasourceschema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: markdown,
		Attributes: map[string]datasourceschema.Attribute{
			"parts": datasourceschema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "Multipart form parts.",
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: dataSourceMultipartPartAttributes(),
				},
			},
		},
	}
}

func actionMultipartPartAttributes() map[string]actionschema.Attribute {
	return map[string]actionschema.Attribute{
		"name": actionschema.StringAttribute{
			Required:            true,
			MarkdownDescription: multipartPartNameDescription,
		},
		"value": actionschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartValueDescription,
		},
		"file_path": actionschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartFilePathDescription,
		},
		"content_type": actionschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartContentTypeDescription,
		},
	}
}

func actionMultipartSchema(markdown string) actionschema.SingleNestedAttribute {
	return actionschema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: markdown,
		Attributes: map[string]actionschema.Attribute{
			"parts": actionschema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "Multipart form parts.",
				NestedObject: actionschema.NestedAttributeObject{
					Attributes: actionMultipartPartAttributes(),
				},
			},
		},
	}
}

func ephemeralMultipartPartAttributes() map[string]ephemeralschema.Attribute {
	return map[string]ephemeralschema.Attribute{
		"name": ephemeralschema.StringAttribute{
			Required:            true,
			MarkdownDescription: multipartPartNameDescription,
		},
		"value": ephemeralschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartValueDescription,
		},
		"file_path": ephemeralschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartFilePathDescription,
		},
		"content_type": ephemeralschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: multipartPartContentTypeDescription,
		},
	}
}

func ephemeralMultipartSchema(markdown string) ephemeralschema.SingleNestedAttribute {
	return ephemeralschema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: markdown,
		Attributes: map[string]ephemeralschema.Attribute{
			"parts": ephemeralschema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "Multipart form parts.",
				NestedObject: ephemeralschema.NestedAttributeObject{
					Attributes: ephemeralMultipartPartAttributes(),
				},
			},
		},
	}
}

func multipartObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: multipartPartAttrTypes}
}
