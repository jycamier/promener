package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetric_FullName(t *testing.T) {
	tests := []struct {
		name      string
		metric    Metric
		want      string
	}{
		{
			name: "full metric with namespace, subsystem and name",
			metric: Metric{
				Namespace: "http",
				Subsystem: "server",
				Name:      "requests_total",
			},
			want: "http_server_requests_total",
		},
		{
			name: "metric with namespace and name only",
			metric: Metric{
				Namespace: "http",
				Name:      "requests_total",
			},
			want: "http_requests_total",
		},
		{
			name: "metric with name only",
			metric: Metric{
				Name: "requests_total",
			},
			want: "requests_total",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metric.FullName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetric_Validate(t *testing.T) {
	tests := []struct {
		name    string
		metric  Metric
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid counter metric",
			metric: Metric{
				Name:      "requests_total",
				Namespace: "http",
				Subsystem: "server",
				Type:      MetricTypeCounter,
				Help:      "Total HTTP requests",
				Labels:    Labels{{Name: "method"}, {Name: "status"}},
			},
			wantErr: false,
		},
		{
			name: "valid histogram metric",
			metric: Metric{
				Name:      "request_duration_seconds",
				Namespace: "http",
				Subsystem: "server",
				Type:      MetricTypeHistogram,
				Help:      "Request duration",
				Labels:    Labels{{Name: "method"}},
				Buckets:   []float64{0.1, 0.5, 1.0},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			metric: Metric{
				Namespace: "http",
				Subsystem: "server",
				Type:      MetricTypeCounter,
				Help:      "Test",
			},
			wantErr: true,
			errMsg:  "metric name is required",
		},
		{
			name: "invalid name format",
			metric: Metric{
				Name:      "123invalid",
				Namespace: "http",
				Subsystem: "server",
				Type:      MetricTypeCounter,
				Help:      "Test",
			},
			wantErr: true,
			errMsg:  "invalid metric name",
		},
		{
			name: "missing namespace",
			metric: Metric{
				Name:      "test",
				Subsystem: "server",
				Type:      MetricTypeCounter,
				Help:      "Test",
			},
			wantErr: true,
			errMsg:  "metric namespace is required",
		},
		{
			name: "missing subsystem",
			metric: Metric{
				Name:      "test",
				Namespace: "http",
				Type:      MetricTypeCounter,
				Help:      "Test",
			},
			wantErr: true,
			errMsg:  "metric subsystem is required",
		},
		{
			name: "invalid metric type",
			metric: Metric{
				Name:      "test",
				Namespace: "http",
				Subsystem: "server",
				Type:      "invalid",
				Help:      "Test",
			},
			wantErr: true,
			errMsg:  "invalid metric type",
		},
		{
			name: "missing help",
			metric: Metric{
				Name:      "test",
				Namespace: "http",
				Subsystem: "server",
				Type:      MetricTypeCounter,
			},
			wantErr: true,
			errMsg:  "metric help is required",
		},
		{
			name: "invalid label name",
			metric: Metric{
				Name:      "test",
				Namespace: "http",
				Subsystem: "server",
				Type:      MetricTypeCounter,
				Help:      "Test",
				Labels:    Labels{{Name: "123invalid"}},
			},
			wantErr: true,
			errMsg:  "invalid label name",
		},
		{
			name: "histogram without buckets",
			metric: Metric{
				Name:      "test",
				Namespace: "http",
				Subsystem: "server",
				Type:      MetricTypeHistogram,
				Help:      "Test",
			},
			wantErr: true,
			errMsg:  "histogram metrics require buckets",
		},
		{
			name: "invalid const label name",
			metric: Metric{
				Name:      "test",
				Namespace: "http",
				Subsystem: "server",
				Type:      MetricTypeCounter,
				Help:      "Test",
				ConstLabels: map[string]string{
					"123invalid": "value",
				},
			},
			wantErr: true,
			errMsg:  "invalid const label name",
		},
		{
			name: "valid const labels",
			metric: Metric{
				Name:      "test",
				Namespace: "http",
				Subsystem: "server",
				Type:      MetricTypeCounter,
				Help:      "Test",
				ConstLabels: map[string]string{
					"environment": "production",
					"version":     "1.0.0",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMetricType_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		metricType MetricType
		want       bool
	}{
		{
			name:       "valid counter",
			metricType: MetricTypeCounter,
			want:       true,
		},
		{
			name:       "valid gauge",
			metricType: MetricTypeGauge,
			want:       true,
		},
		{
			name:       "valid histogram",
			metricType: MetricTypeHistogram,
			want:       true,
		},
		{
			name:       "valid summary",
			metricType: MetricTypeSummary,
			want:       true,
		},
		{
			name:       "invalid type",
			metricType: "invalid",
			want:       false,
		},
		{
			name:       "empty type",
			metricType: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metricType.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}
