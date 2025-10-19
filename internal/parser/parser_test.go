package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jycamier/promener/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
		check   func(t *testing.T, spec *domain.Specification)
	}{
		{
			name: "valid specification",
			yaml: `
version: "1.0"
info:
  title: "Test Metrics"
  version: "1.0.0"

services:
  default:
    info:
      title: "Test Service"
      version: "1.0.0"
      package: "metrics"
    metrics:
      requests_total:
        namespace: http
        subsystem: server
        type: counter
        help: "Total requests"
        labels:
          - method
          - status
`,
			wantErr: false,
			check: func(t *testing.T, spec *domain.Specification) {
				assert.Equal(t, "1.0", spec.Version)
				assert.Equal(t, "Test Metrics", spec.Info.Title)
				require.Len(t, spec.Services, 1)

				service, ok := spec.Services["default"]
				require.True(t, ok)
				require.Len(t, service.Metrics, 1)

				metric, ok := service.Metrics["requests_total"]
				require.True(t, ok)
				assert.Equal(t, "http", metric.Namespace)
				assert.Equal(t, "server", metric.Subsystem)
				assert.Equal(t, domain.MetricTypeCounter, metric.Type)
				assert.Equal(t, "Total requests", metric.Help)
				assert.Equal(t, []string{"method", "status"}, metric.GetLabelNames())
			},
		},
		{
			name: "metric with const labels",
			yaml: `
version: "1.0"
info:
  title: "Test Metrics"
  version: "1.0.0"

services:
  default:
    info:
      title: "Test Service"
      version: "1.0.0"
      package: "metrics"
    metrics:
      requests_total:
        namespace: http
        subsystem: server
        type: counter
        help: "Total requests"
        labels:
          - method
        constLabels:
          environment: "production"
          version: "1.0.0"
`,
			wantErr: false,
			check: func(t *testing.T, spec *domain.Specification) {
				metric := spec.Services["default"].Metrics["requests_total"]
				require.NotNil(t, metric.ConstLabels)
				constLabelsMap := metric.ConstLabels.ToMap()
				assert.Equal(t, "production", constLabelsMap["environment"])
				assert.Equal(t, "1.0.0", constLabelsMap["version"])
			},
		},
		{
			name: "histogram with buckets",
			yaml: `
version: "1.0"
info:
  title: "Test Metrics"
  version: "1.0.0"

services:
  default:
    info:
      title: "Test Service"
      version: "1.0.0"
      package: "metrics"
    metrics:
      request_duration:
        namespace: http
        subsystem: server
        type: histogram
        help: "Request duration"
        labels:
          - method
        buckets: [0.1, 0.5, 1.0, 5.0]
`,
			wantErr: false,
			check: func(t *testing.T, spec *domain.Specification) {
				metric := spec.Services["default"].Metrics["request_duration"]
				assert.Equal(t, domain.MetricTypeHistogram, metric.Type)
				assert.Equal(t, []float64{0.1, 0.5, 1.0, 5.0}, metric.Buckets)
			},
		},
		{
			name: "summary with objectives",
			yaml: `
version: "1.0"
info:
  title: "Test Metrics"
  version: "1.0.0"

services:
  default:
    info:
      title: "Test Service"
      version: "1.0.0"
      package: "metrics"
    metrics:
      response_size:
        namespace: http
        subsystem: server
        type: summary
        help: "Response size"
        labels:
          - method
        objectives:
          0.5: 0.05
          0.9: 0.01
          0.99: 0.001
`,
			wantErr: false,
			check: func(t *testing.T, spec *domain.Specification) {
				metric := spec.Services["default"].Metrics["response_size"]
				assert.Equal(t, domain.MetricTypeSummary, metric.Type)
				require.NotNil(t, metric.Objectives)
				assert.Equal(t, 0.05, metric.Objectives[0.5])
				assert.Equal(t, 0.01, metric.Objectives[0.9])
				assert.Equal(t, 0.001, metric.Objectives[0.99])
			},
		},
		{
			name: "invalid yaml",
			yaml: `
invalid: yaml: syntax
`,
			wantErr: true,
			errMsg:  "failed to unmarshal YAML",
		},
		{
			name: "missing required fields",
			yaml: `
version: "1.0"
info:
  title: "Test Metrics"

services:
  default:
    info:
      title: "Test Service"
    metrics:
      test:
        type: counter
`,
			wantErr: true,
			errMsg:  "invalid specification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			spec, err := p.Parse([]byte(tt.yaml))

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, spec)
				if tt.check != nil {
					tt.check(t, spec)
				}
			}
		})
	}
}

func TestParser_ParseFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("parse valid file", func(t *testing.T) {
		yamlContent := `
version: "1.0"
info:
  title: "Test Metrics"
  version: "1.0.0"

services:
  default:
    info:
      title: "Test Service"
      version: "1.0.0"
      package: "metrics"
    metrics:
      test_metric:
        namespace: test
        subsystem: example
        type: counter
        help: "Test metric"
        labels:
          - label1
`
		filePath := filepath.Join(tempDir, "valid.yaml")
		err := os.WriteFile(filePath, []byte(yamlContent), 0644)
		require.NoError(t, err)

		p := New()
		spec, err := p.ParseFile(filePath)
		require.NoError(t, err)
		require.NotNil(t, spec)
		assert.Equal(t, "Test Metrics", spec.Info.Title)
	})

	t.Run("file not found", func(t *testing.T) {
		p := New()
		_, err := p.ParseFile(filepath.Join(tempDir, "nonexistent.yaml"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})

	t.Run("invalid yaml in file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "invalid.yaml")
		err := os.WriteFile(filePath, []byte("invalid: yaml: content:"), 0644)
		require.NoError(t, err)

		p := New()
		_, err = p.ParseFile(filePath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal YAML")
	})
}

func TestParser_EnrichMetricsWithNames(t *testing.T) {
	yaml := `
version: "1.0"
info:
  title: "Test Metrics"
  version: "1.0.0"

services:
  default:
    info:
      title: "Test Service"
      version: "1.0.0"
      package: "metrics"
    metrics:
      my_metric:
        namespace: test
        subsystem: example
        type: counter
        help: "Test metric"
`

	p := New()
	spec, err := p.Parse([]byte(yaml))
	require.NoError(t, err)

	metric := spec.Services["default"].Metrics["my_metric"]
	assert.Equal(t, "my_metric", metric.Name)
}
