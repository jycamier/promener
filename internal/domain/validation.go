package domain

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

// Validation represents a compiled CEL validation expression
type Validation struct {
	Expression string
	Program    cel.Program
}

// ParseValidation compiles a CEL expression for label validation
// The expression has access to a 'value' variable containing the label value
// Examples:
//   - value.matches('^mon-service$')
//   - value in ['service1', 'service2']
//   - value.startsWith('prod-')
//   - size(value) > 3
func ParseValidation(expression string) (*Validation, error) {
	// Create CEL environment with 'value' variable as string
	env, err := cel.NewEnv(
		cel.Variable("value", cel.StringType),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	// Compile the expression
	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("failed to compile CEL expression '%s': %w", expression, issues.Err())
	}

	// Check that the expression returns a boolean
	if ast.OutputType() != cel.BoolType {
		return nil, fmt.Errorf("CEL expression must return a boolean, got %s", ast.OutputType())
	}

	// Create the program
	program, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL program: %w", err)
	}

	return &Validation{
		Expression: expression,
		Program:    program,
	}, nil
}

// Validate checks if a value passes the CEL validation expression
func (v *Validation) Validate(value string) error {
	// Evaluate the expression with the value
	result, _, err := v.Program.Eval(map[string]interface{}{
		"value": value,
	})
	if err != nil {
		return fmt.Errorf("failed to evaluate CEL expression: %w", err)
	}

	// Check if result is a boolean
	boolResult, ok := result.(types.Bool)
	if !ok {
		return fmt.Errorf("CEL expression did not return a boolean: %v", result)
	}

	// If false, validation failed
	if boolResult == types.False {
		return fmt.Errorf("value '%s' failed validation: %s", value, v.Expression)
	}

	return nil
}
