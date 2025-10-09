package rules

import (
	"fmt"
	"path/filepath"

	"github.com/arsiba/tofulint-ruleset-opentofu/project"
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

const (
	filenameMain          = "main.tf"
	filenameTofuMain      = "main.tofu"
	filenameVariables     = "variables.tf"
	filenameTofuVariables = "variables.tofu"
	filenameOutputs       = "outputs.tf"
	filenameTofuOutputs   = "outputs.tofu"
)

// TerraformStandardModuleStructureRule checks whether modules adhere to Terraform's standard module structure
type TerraformStandardModuleStructureRule struct {
	tflint.DefaultRule
}

// NewTerraformStandardModuleStructureRule returns a new rule
func NewTerraformStandardModuleStructureRule() *TerraformStandardModuleStructureRule {
	return &TerraformStandardModuleStructureRule{}
}

// Name returns the rule name
func (r *TerraformStandardModuleStructureRule) Name() string {
	return "terraform_standard_module_structure"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformStandardModuleStructureRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformStandardModuleStructureRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformStandardModuleStructureRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check emits errors for any missing files and any block types that are included in the wrong file
func (r *TerraformStandardModuleStructureRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		// This rule does not run on non-Terraform directory.
		return nil
	}

	body, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
				Body:       &hclext.BodySchema{},
			},
			{
				Type:       "output",
				LabelNames: []string{"name"},
				Body:       &hclext.BodySchema{},
			},
		},
	}, &tflint.GetModuleContentOption{ExpandMode: tflint.ExpandModeNone})
	if err != nil {
		return err
	}

	blocks := body.Blocks.ByType()

	if err := r.checkFiles(runner, body.Blocks); err != nil {
		return err
	}
	if err := r.checkVariables(runner, blocks["variable"]); err != nil {
		return err
	}
	if err := r.checkOutputs(runner, blocks["output"]); err != nil {
		return err
	}

	return nil
}

func (r *TerraformStandardModuleStructureRule) checkFiles(runner tflint.Runner, blocks hclext.Blocks) error {
	onlyJSON, err := r.onlyJSON(runner)
	if err != nil {
		return err
	}
	if onlyJSON {
		return nil
	}

	f, err := runner.GetFiles()
	if err != nil {
		return err
	}

	var dir string
	files := make(map[string]*hcl.File, len(f))
	for name, file := range f {
		dir = filepath.Dir(name)
		files[filepath.Base(name)] = file
	}

	if files[filenameTofuMain] == nil && files[filenameMain] == nil {
		if err := runner.EmitIssue(
			r,
			fmt.Sprintf("Module should include a %s or %s file as the primary entrypoint", filenameTofuMain, filenameMain),
			hcl.Range{
				Filename: filepath.Join(dir, filenameTofuMain),
				Start:    hcl.InitialPos,
			},
		); err != nil {
			return err
		}
	}

	if files[filenameVariables] == nil && files[filenameTofuVariables] == nil && len(blocks.ByType()["variable"]) == 0 {
		if err := runner.EmitIssue(
			r,
			fmt.Sprintf("Module should include an empty %s or %s file", filenameVariables, filenameTofuVariables),
			hcl.Range{
				Filename: filepath.Join(dir, filenameTofuVariables),
				Start:    hcl.InitialPos,
			},
		); err != nil {
			return err
		}
	}

	if files[filenameTofuOutputs] == nil && files[filenameOutputs] == nil && len(blocks.ByType()["output"]) == 0 {
		if err := runner.EmitIssue(
			r,
			fmt.Sprintf("Module should include an empty %s or %s file", filenameTofuOutputs, filenameOutputs),
			hcl.Range{
				Filename: filepath.Join(dir, filenameTofuOutputs),
				Start:    hcl.InitialPos,
			},
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformStandardModuleStructureRule) checkVariables(runner tflint.Runner, variables hclext.Blocks) error {
	for _, variable := range variables {
		filename := variable.DefRange.Filename
		if r.shouldMove(filename, filenameTofuVariables) || r.shouldMove(filename, filenameVariables) {
			target := filenameVariables
			if r.shouldMove(filename, filenameTofuVariables) {
				target = filenameTofuVariables
			}

			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("variable %q should be moved from %s to %s", variable.Labels[0], filename, target),
				variable.DefRange,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformStandardModuleStructureRule) checkOutputs(runner tflint.Runner, outputs hclext.Blocks) error {
	for _, output := range outputs {
		filename := output.DefRange.Filename
		if r.shouldMove(filename, filenameTofuOutputs) || r.shouldMove(filename, filenameOutputs) {
			target := filenameOutputs
			if r.shouldMove(filename, filenameTofuOutputs) {
				target = filenameTofuOutputs
			}

			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("output %q should be moved from %s to %s", output.Labels[0], filename, target),
				output.DefRange,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformStandardModuleStructureRule) onlyJSON(runner tflint.Runner) (bool, error) {
	files, err := runner.GetFiles()
	if err != nil {
		return false, err
	}

	if len(files) == 0 {
		return false, nil
	}

	for filename := range files {
		if filepath.Ext(filename) != ".json" {
			return false, nil
		}
	}

	return true, nil
}

func (r *TerraformStandardModuleStructureRule) shouldMove(path string, expected string) bool {
	// json files are likely generated and conventional filenames do not apply
	if filepath.Ext(path) == ".json" {
		return false
	}

	return filepath.Base(path) != expected
}
