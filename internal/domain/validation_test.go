package domain

import (
	"fmt"
	"testing"
)

// validateLabelValue validates a label value against all its validation rules (test helper)
func validateLabelValue(label *LabelDefinition, value string) error {
	for _, validationExpr := range label.Validations {
		validation, err := ParseValidation(validationExpr)
		if err != nil {
			return fmt.Errorf("invalid validation expression for label '%s': %w", label.Name, err)
		}

		if err := validation.Validate(value); err != nil {
			return fmt.Errorf("validation failed for label '%s': %w", label.Name, err)
		}
	}
	return nil
}

func TestParseValidation(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name:    "regexp matches",
			expr:    "value.matches('^mon-service$')",
			wantErr: false,
		},
		{
			name:    "in list",
			expr:    "value in ['service1', 'service2', 'service3']",
			wantErr: false,
		},
		{
			name:    "startsWith",
			expr:    "value.startsWith('prod-')",
			wantErr: false,
		},
		{
			name:    "endsWith",
			expr:    "value.endsWith('-svc')",
			wantErr: false,
		},
		{
			name:    "size check",
			expr:    "size(value) > 3",
			wantErr: false,
		},
		{
			name:    "complex expression",
			expr:    "value.matches('^[a-z]+$') && size(value) >= 3",
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			expr:    "value.matches(",
			wantErr: true,
		},
		{
			name:    "non-boolean expression",
			expr:    "size(value)",
			wantErr: true,
		},
		{
			name:    "undefined variable",
			expr:    "foo == 'bar'",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseValidation(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseValidation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Expression != tt.expr {
				t.Errorf("ParseValidation() Expression = %v, want %v", got.Expression, tt.expr)
			}
			if got.Program == nil {
				t.Errorf("ParseValidation() Program is nil")
			}
		})
	}
}

func TestValidation_Validate(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		testValue string
		wantErr   bool
	}{
		{
			name:      "matches - pass",
			expr:      "value.matches('^mon-service$')",
			testValue: "mon-service",
			wantErr:   false,
		},
		{
			name:      "matches - fail",
			expr:      "value.matches('^mon-service$')",
			testValue: "autre-service",
			wantErr:   true,
		},
		{
			name:      "in list - pass",
			expr:      "value in ['service1', 'service2']",
			testValue: "service1",
			wantErr:   false,
		},
		{
			name:      "in list - fail",
			expr:      "value in ['service1', 'service2']",
			testValue: "service3",
			wantErr:   true,
		},
		{
			name:      "startsWith - pass",
			expr:      "value.startsWith('prod-')",
			testValue: "prod-api",
			wantErr:   false,
		},
		{
			name:      "startsWith - fail",
			expr:      "value.startsWith('prod-')",
			testValue: "dev-api",
			wantErr:   true,
		},
		{
			name:      "size check - pass",
			expr:      "size(value) > 3",
			testValue: "test",
			wantErr:   false,
		},
		{
			name:      "size check - fail",
			expr:      "size(value) > 3",
			testValue: "ab",
			wantErr:   true,
		},
		{
			name:      "complex - pass",
			expr:      "value.matches('^[a-z]+$') && size(value) >= 3",
			testValue: "abc",
			wantErr:   false,
		},
		{
			name:      "complex - fail pattern",
			expr:      "value.matches('^[a-z]+$') && size(value) >= 3",
			testValue: "ABC",
			wantErr:   true,
		},
		{
			name:      "complex - fail size",
			expr:      "value.matches('^[a-z]+$') && size(value) >= 3",
			testValue: "ab",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validation, err := ParseValidation(tt.expr)
			if err != nil {
				t.Fatalf("Failed to parse validation: %v", err)
			}

			err = validation.Validate(tt.testValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validation.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLabelValue(t *testing.T) {
	tests := []struct {
		name      string
		label     *LabelDefinition
		value     string
		wantErr   bool
	}{
		{
			name: "single validation pass",
			label: &LabelDefinition{
				Name: "service",
				Validations: []string{
					"value.matches('^[a-z-]+$')",
				},
			},
			value:   "my-service",
			wantErr: false,
		},
		{
			name: "single validation fail",
			label: &LabelDefinition{
				Name: "service",
				Validations: []string{
					"value.matches('^[a-z-]+$')",
				},
			},
			value:   "MyService",
			wantErr: true,
		},
		{
			name: "multiple validations all pass",
			label: &LabelDefinition{
				Name: "service",
				Validations: []string{
					"value.matches('^[a-z-]+$')",
					"size(value) >= 3",
					"value.startsWith('svc-')",
				},
			},
			value:   "svc-api",
			wantErr: false,
		},
		{
			name: "multiple validations one fails",
			label: &LabelDefinition{
				Name: "service",
				Validations: []string{
					"value.matches('^[a-z-]+$')",
					"size(value) >= 3",
					"value.startsWith('svc-')",
				},
			},
			value:   "api", // doesn't start with 'svc-'
			wantErr: true,
		},
		{
			name: "no validations",
			label: &LabelDefinition{
				Name:        "service",
				Validations: []string{},
			},
			value:   "anything-goes",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLabelValue(tt.label, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLabelValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
