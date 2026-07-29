package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func writeOnlyHeadersSchema(markdown string) schema.MapAttribute {
	return schema.MapAttribute{
		ElementType:         types.StringType,
		Optional:            true,
		Sensitive:           true,
		WriteOnly:           true,
		MarkdownDescription: markdown,
	}
}

func writeOnlyBodySchema(markdown string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:            true,
		Sensitive:           true,
		WriteOnly:           true,
		MarkdownDescription: markdown,
	}
}

func writeOnlyVersionSchema(markdown string) schema.Int64Attribute {
	return schema.Int64Attribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: markdown,
		Default:             int64default.StaticInt64(0),
	}
}
