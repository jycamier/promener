package validator

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

// RegoValidator validates specifications using Rego rules.
type RegoValidator struct {
	rulesDirs []string
}

// NewRegoValidator creates a new Rego validator.
func NewRegoValidator(rulesDirs []string) *RegoValidator {
	return &RegoValidator{
		rulesDirs: rulesDirs,
	}
}

// Validate runs Rego rules against the provided specification.
func (v *RegoValidator) Validate(ctx context.Context, input interface{}) ([]ValidationError, error) {
	if len(v.rulesDirs) == 0 {
		return nil, nil
	}

	// Use RuleSourceResolver to load rules from all sources (local, HTTP, Git)
	resolver := NewRuleSourceResolver()

	var regoFiles []string
	for _, source := range v.rulesDirs {
		files, err := resolver.Load(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("failed to load rules from %s: %w", source, err)
		}
		regoFiles = append(regoFiles, files...)
	}

	if len(regoFiles) == 0 {
		return nil, nil
	}

	// Prepare Rego evaluation with Rego v1 syntax support
	query, err := rego.New(
		rego.Query("data.PromenerPolicy.PromenerPolicy"),
		rego.Load(regoFiles, nil),
		rego.SetRegoVersion(ast.RegoV1),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare rego query: %w", err)
	}

	// Evaluate rules
	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate rego rules: %w", err)
	}

	var validationErrors []ValidationError
	for _, result := range results {
		for _, expr := range result.Expressions {
			// Simplified style using 'path' and 'message'
			if val, ok := expr.Value.([]interface{}); ok {
				for _, item := range val {
					if res, ok := item.(map[string]interface{}); ok {
						msg := ""
						if m, ok := res["message"].(string); ok {
							msg = m
						}

						path := ""
						if p, ok := res["path"].(string); ok {
							path = p
						}

						severity := "error"
						if s, ok := res["severity"].(string); ok {
							severity = s
						}

						if msg != "" {
							validationErrors = append(validationErrors, ValidationError{
								Path:     path,
								Message:  msg,
								Source:   "rego",
								Severity: severity,
							})
						}
					}
				}
			}
		}
	}

	return validationErrors, nil
}
