package main

import (
	"github.com/arsiba/tofulint-ruleset-opentofu/project"
	"github.com/arsiba/tofulint-ruleset-opentofu/rules"
	"github.com/arsiba/tofulint-ruleset-opentofu/terraform"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &terraform.RuleSet{
			BuiltinRuleSet: tflint.BuiltinRuleSet{
				Name:       "opentofu",
				Version:    project.Version,
				Constraint: ">= 0.0.1",
			},
			PresetRules: rules.PresetRules,
		},
	})
}
