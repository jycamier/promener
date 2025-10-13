package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEnvVarValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  EnvVarValue
	}{
		{
			name:  "literal value",
			value: "production",
			want: EnvVarValue{
				IsEnvVar:     false,
				LiteralValue: "production",
			},
		},
		{
			name:  "env var without default",
			value: "${ENVIRONMENT}",
			want: EnvVarValue{
				IsEnvVar:     true,
				EnvVar:       "ENVIRONMENT",
				DefaultValue: "",
			},
		},
		{
			name:  "env var with default",
			value: "${ENVIRONMENT:production}",
			want: EnvVarValue{
				IsEnvVar:     true,
				EnvVar:       "ENVIRONMENT",
				DefaultValue: "production",
			},
		},
		{
			name:  "env var with empty default",
			value: "${ENVIRONMENT:}",
			want: EnvVarValue{
				IsEnvVar:     true,
				EnvVar:       "ENVIRONMENT",
				DefaultValue: "",
			},
		},
		{
			name:  "env var with complex default",
			value: "${REGION:us-east-1}",
			want: EnvVarValue{
				IsEnvVar:     true,
				EnvVar:       "REGION",
				DefaultValue: "us-east-1",
			},
		},
		{
			name:  "not an env var pattern",
			value: "some-value-with-${}",
			want: EnvVarValue{
				IsEnvVar:     false,
				LiteralValue: "some-value-with-${}",
			},
		},
		{
			name:  "env var with underscores",
			value: "${MY_CUSTOM_VAR}",
			want: EnvVarValue{
				IsEnvVar:     true,
				EnvVar:       "MY_CUSTOM_VAR",
				DefaultValue: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseEnvVarValue(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEnvVarValue_ToGoCode(t *testing.T) {
	tests := []struct {
		name  string
		value EnvVarValue
		want  string
	}{
		{
			name: "literal value",
			value: EnvVarValue{
				IsEnvVar:     false,
				LiteralValue: "production",
			},
			want: `"production"`,
		},
		{
			name: "env var without default",
			value: EnvVarValue{
				IsEnvVar: true,
				EnvVar:   "ENVIRONMENT",
			},
			want: `os.Getenv("ENVIRONMENT")`,
		},
		{
			name: "env var with default",
			value: EnvVarValue{
				IsEnvVar:     true,
				EnvVar:       "ENVIRONMENT",
				DefaultValue: "production",
			},
			want: `getEnvOrDefault("ENVIRONMENT", "production")`,
		},
		{
			name: "env var with complex default",
			value: EnvVarValue{
				IsEnvVar:     true,
				EnvVar:       "AWS_REGION",
				DefaultValue: "us-east-1",
			},
			want: `getEnvOrDefault("AWS_REGION", "us-east-1")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.ToGoCode()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEnvVarValue_NeedsOsImport(t *testing.T) {
	tests := []struct {
		name  string
		value EnvVarValue
		want  bool
	}{
		{
			name: "literal value does not need os import",
			value: EnvVarValue{
				IsEnvVar:     false,
				LiteralValue: "production",
			},
			want: false,
		},
		{
			name: "env var needs os import",
			value: EnvVarValue{
				IsEnvVar: true,
				EnvVar:   "ENVIRONMENT",
			},
			want: true,
		},
		{
			name: "env var with default needs os import",
			value: EnvVarValue{
				IsEnvVar:     true,
				EnvVar:       "ENVIRONMENT",
				DefaultValue: "production",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.NeedsOsImport()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEnvVarValue_NeedsHelperFunction(t *testing.T) {
	tests := []struct {
		name  string
		value EnvVarValue
		want  bool
	}{
		{
			name: "literal value does not need helper",
			value: EnvVarValue{
				IsEnvVar:     false,
				LiteralValue: "production",
			},
			want: false,
		},
		{
			name: "env var without default does not need helper",
			value: EnvVarValue{
				IsEnvVar: true,
				EnvVar:   "ENVIRONMENT",
			},
			want: false,
		},
		{
			name: "env var with default needs helper",
			value: EnvVarValue{
				IsEnvVar:     true,
				EnvVar:       "ENVIRONMENT",
				DefaultValue: "production",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.NeedsHelperFunction()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseConstLabels(t *testing.T) {
	constLabels := map[string]string{
		"environment": "${ENVIRONMENT:production}",
		"region":      "${REGION}",
		"version":     "1.0.0",
	}

	result := ParseConstLabels(constLabels)

	assert.Len(t, result, 3)

	// Check environment
	assert.True(t, result["environment"].IsEnvVar)
	assert.Equal(t, "ENVIRONMENT", result["environment"].EnvVar)
	assert.Equal(t, "production", result["environment"].DefaultValue)

	// Check region
	assert.True(t, result["region"].IsEnvVar)
	assert.Equal(t, "REGION", result["region"].EnvVar)
	assert.Equal(t, "", result["region"].DefaultValue)

	// Check version
	assert.False(t, result["version"].IsEnvVar)
	assert.Equal(t, "1.0.0", result["version"].LiteralValue)
}

func TestHasEnvVars(t *testing.T) {
	tests := []struct {
		name        string
		constLabels map[string]string
		want        bool
	}{
		{
			name: "has env vars",
			constLabels: map[string]string{
				"environment": "${ENVIRONMENT}",
				"version":     "1.0.0",
			},
			want: true,
		},
		{
			name: "no env vars",
			constLabels: map[string]string{
				"environment": "production",
				"version":     "1.0.0",
			},
			want: false,
		},
		{
			name:        "empty const labels",
			constLabels: map[string]string{},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasEnvVars(tt.constLabels)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNeedsHelperFunc(t *testing.T) {
	tests := []struct {
		name        string
		constLabels map[string]string
		want        bool
	}{
		{
			name: "needs helper func",
			constLabels: map[string]string{
				"environment": "${ENVIRONMENT:production}",
				"version":     "1.0.0",
			},
			want: true,
		},
		{
			name: "does not need helper func",
			constLabels: map[string]string{
				"environment": "${ENVIRONMENT}",
				"version":     "1.0.0",
			},
			want: false,
		},
		{
			name: "no env vars",
			constLabels: map[string]string{
				"environment": "production",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsHelperFunc(tt.constLabels)
			assert.Equal(t, tt.want, got)
		})
	}
}
