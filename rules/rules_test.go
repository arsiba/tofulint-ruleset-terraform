package rules

import (
	"testing"

	"github.com/arsiba/tofulint-ruleset-terraform/terraform"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func testRunner(t *testing.T, files map[string]string) *terraform.Runner {
	return terraform.NewRunner(helper.TestRunner(t, files))
}
