package provider

import (
	actionvalidator "github.com/hashicorp/terraform-plugin-framework-validators/actionvalidator"
	datasourcevalidator "github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	ephemeralvalidator "github.com/hashicorp/terraform-plugin-framework-validators/ephemeralvalidator"
	resourcevalidator "github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func resourceBodyValidators(body, bodyWo, bodyFile, bodyMultipart string) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(path.MatchRoot(body), path.MatchRoot(bodyWo)),
		resourcevalidator.Conflicting(path.MatchRoot(body), path.MatchRoot(bodyFile)),
		resourcevalidator.Conflicting(path.MatchRoot(body), path.MatchRoot(bodyMultipart)),
		resourcevalidator.Conflicting(path.MatchRoot(bodyWo), path.MatchRoot(bodyFile)),
		resourcevalidator.Conflicting(path.MatchRoot(bodyWo), path.MatchRoot(bodyMultipart)),
		resourcevalidator.Conflicting(path.MatchRoot(bodyFile), path.MatchRoot(bodyMultipart)),
	}
}

func dataSourceBodyValidators(body, bodyFile, bodyMultipart string) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(path.MatchRoot(body), path.MatchRoot(bodyFile)),
		datasourcevalidator.Conflicting(path.MatchRoot(body), path.MatchRoot(bodyMultipart)),
		datasourcevalidator.Conflicting(path.MatchRoot(bodyFile), path.MatchRoot(bodyMultipart)),
	}
}

func actionBodyValidators(body, bodyFile, bodyMultipart string) []action.ConfigValidator {
	return []action.ConfigValidator{
		actionvalidator.Conflicting(path.MatchRoot(body), path.MatchRoot(bodyFile)),
		actionvalidator.Conflicting(path.MatchRoot(body), path.MatchRoot(bodyMultipart)),
		actionvalidator.Conflicting(path.MatchRoot(bodyFile), path.MatchRoot(bodyMultipart)),
	}
}

func ephemeralBodyValidators(body, bodyFile, bodyMultipart string) []ephemeral.ConfigValidator {
	return []ephemeral.ConfigValidator{
		ephemeralvalidator.Conflicting(path.MatchRoot(body), path.MatchRoot(bodyFile)),
		ephemeralvalidator.Conflicting(path.MatchRoot(body), path.MatchRoot(bodyMultipart)),
		ephemeralvalidator.Conflicting(path.MatchRoot(bodyFile), path.MatchRoot(bodyMultipart)),
	}
}
