package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLabelDefinition_IsInherited(t *testing.T) {
	tests := []struct {
		name  string
		label LabelDefinition
		want  bool
	}{
		{
			name: "inherited label with text",
			label: LabelDefinition{
				Name:      "cluster",
				Inherited: "This label is added by Prometheus relabeling",
			},
			want: true,
		},
		{
			name: "non-inherited label",
			label: LabelDefinition{
				Name: "method",
			},
			want: false,
		},
		{
			name: "label with empty inherited field",
			label: LabelDefinition{
				Name:      "status",
				Inherited: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.label.IsInherited()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabels_NonInheritedLabels(t *testing.T) {
	labels := Labels{
		{Name: "method", Description: "HTTP method"},
		{Name: "status", Description: "HTTP status"},
		{Name: "cluster", Inherited: "Added by relabeling"},
		{Name: "region", Inherited: "Added by relabeling"},
	}

	nonInherited := labels.NonInheritedLabels()

	assert.Len(t, nonInherited, 2)
	assert.Equal(t, "method", nonInherited[0].Name)
	assert.Equal(t, "status", nonInherited[1].Name)
}

func TestLabels_InheritedLabels(t *testing.T) {
	labels := Labels{
		{Name: "method", Description: "HTTP method"},
		{Name: "status", Description: "HTTP status"},
		{Name: "cluster", Inherited: "Added by relabeling"},
		{Name: "region", Inherited: "Added by relabeling"},
	}

	inherited := labels.InheritedLabels()

	assert.Len(t, inherited, 2)
	assert.Equal(t, "cluster", inherited[0].Name)
	assert.Equal(t, "region", inherited[1].Name)
}

func TestLabels_UnmarshalYAML_WithInherited(t *testing.T) {
	yamlData := `
method:
  description: "HTTP method"
status:
  description: "HTTP status code"
cluster:
  description: "Kubernetes cluster name"
  inherited: "This label is added via Prometheus relabeling based on the service discovery metadata"
region:
  description: "Cloud region"
  inherited: "Injected by infrastructure relabeling rules"
  validations:
    - "value.matches('^[a-z]+-[a-z]+-[0-9]+$')"
`

	var labels Labels
	err := yaml.Unmarshal([]byte(yamlData), &labels)
	require.NoError(t, err)

	assert.Len(t, labels, 4)

	// Check non-inherited labels
	nonInherited := labels.NonInheritedLabels()
	assert.Len(t, nonInherited, 2)
	assert.Equal(t, "method", nonInherited[0].Name)
	assert.Equal(t, "HTTP method", nonInherited[0].Description)
	assert.Empty(t, nonInherited[0].Inherited)

	// Check inherited labels
	inherited := labels.InheritedLabels()
	assert.Len(t, inherited, 2)
	assert.Equal(t, "cluster", inherited[0].Name)
	assert.Equal(t, "Kubernetes cluster name", inherited[0].Description)
	assert.Equal(t, "This label is added via Prometheus relabeling based on the service discovery metadata", inherited[0].Inherited)

	assert.Equal(t, "region", inherited[1].Name)
	assert.Equal(t, "Injected by infrastructure relabeling rules", inherited[1].Inherited)
	assert.Len(t, inherited[1].Validations, 1)
}

func TestLabels_ToStringSlice(t *testing.T) {
	labels := Labels{
		{Name: "method"},
		{Name: "status"},
		{Name: "cluster", Inherited: "Added by relabeling"},
	}

	names := labels.ToStringSlice()

	assert.Len(t, names, 3)
	assert.Equal(t, []string{"method", "status", "cluster"}, names)
}
