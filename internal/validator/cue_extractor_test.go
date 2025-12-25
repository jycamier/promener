package validator

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"github.com/jycamier/promener/internal/domain"
)

func TestCueExtractor_Extract(t *testing.T) {
	tests := []struct {
		name       string
		cueContent string
		wantErr    bool
		checks     func(*testing.T, *domain.Specification)
	}{
		{
			name: "valid specification with simple labels",
			cueContent: `
version: "1.0.0"
info: {
	title: "Test Metrics"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title: "Default Service"
			version: "1.0.0"
		}
		metrics: {
			requests_total: {
				namespace: "http"
				subsystem: "server"
				type: "counter"
				help: "Total requests"
				labels: ["method", "status"]
			}
		}
	}
}`,
			wantErr: false,
			checks: func(t *testing.T, spec *domain.Specification) {
				if spec.Info.Title != "Test Metrics" {
					t.Errorf("Info.Title = %q, want %q", spec.Info.Title, "Test Metrics")
				}

				if len(spec.Services) != 1 {
					t.Errorf("len(Services) = %d, want 1", len(spec.Services))
				}

				service, ok := spec.Services["default"]
				if !ok {
					t.Fatal("Service 'default' not found")
				}

				metric, ok := service.Metrics["requests_total"]
				if !ok {
					t.Fatal("Metric 'requests_total' not found")
				}

				if metric.Type != domain.MetricTypeCounter {
					t.Errorf("Metric type = %q, want %q", metric.Type, domain.MetricTypeCounter)
				}

				if len(metric.Labels) != 2 {
					t.Errorf("len(Labels) = %d, want 2", len(metric.Labels))
				}
			},
		},
		{
			name: "metric name enrichment from map key",
			cueContent: `
version: "1.0.0"
info: {
	title: "Test"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title: "Default Service"
			version: "1.0.0"
		}
		metrics: {
			test_metric: {
				namespace: "app"
				subsystem: "test"
				type: "gauge"
				help: "Test gauge"
			}
		}
	}
}`,
			wantErr: false,
			checks: func(t *testing.T, spec *domain.Specification) {
				service := spec.Services["default"]
				metric := service.Metrics["test_metric"]

				// Metric name should be enriched from map key
				if metric.Name != "test_metric" {
					t.Errorf("Metric.Name = %q, want %q", metric.Name, "test_metric")
				}

				// Full name should be namespace_subsystem_name
				if metric.FullName() != "app_test_test_metric" {
					t.Errorf("Metric.FullName() = %q, want %q", metric.FullName(), "app_test_test_metric")
				}
			},
		},
		{
			name: "explicit name overrides map key",
			cueContent: `
version: "1.0.0"
info: {
	title: "Test"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title: "Default Service"
			version: "1.0.0"
		}
		metrics: {
			my_cue_key: {
				name: "requests_total"
				namespace: "http"
				subsystem: "server"
				type: "counter"
				help: "Total HTTP requests"
			}
		}
	}
}`,
			wantErr: false,
			checks: func(t *testing.T, spec *domain.Specification) {
				service := spec.Services["default"]
				metric := service.Metrics["my_cue_key"]

				// Explicit name should be preserved, not overwritten by map key
				if metric.Name != "requests_total" {
					t.Errorf("Metric.Name = %q, want %q", metric.Name, "requests_total")
				}

				// Full name should use the explicit name
				if metric.FullName() != "http_server_requests_total" {
					t.Errorf("Metric.FullName() = %q, want %q", metric.FullName(), "http_server_requests_total")
				}
			},
		},
		{
			name: "detailed labels with descriptions",
			cueContent: `
version: "1.0.0"
info: {
	title: "Test"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title: "Default Service"
			version: "1.0.0"
		}
		metrics: {
			http_requests: {
				namespace: "http"
				subsystem: "server"
				type: "counter"
				help: "HTTP requests"
				labels: {
					method: {
						description: "HTTP method"
					}
					path: {
						description: "Request path"
					}
				}
			}
		}
	}
}`,
			wantErr: false,
			checks: func(t *testing.T, spec *domain.Specification) {
				service := spec.Services["default"]
				metric := service.Metrics["http_requests"]

				if len(metric.Labels) != 2 {
					t.Fatalf("len(Labels) = %d, want 2", len(metric.Labels))
				}

				// Check label descriptions
				foundMethod := false
				foundPath := false
				for _, label := range metric.Labels {
					if label.Name == "method" {
						foundMethod = true
						if label.Description != "HTTP method" {
							t.Errorf("method label description = %q, want %q", label.Description, "HTTP method")
						}
					}
					if label.Name == "path" {
						foundPath = true
						if label.Description != "Request path" {
							t.Errorf("path label description = %q, want %q", label.Description, "Request path")
						}
					}
				}

				if !foundMethod {
					t.Error("method label not found")
				}
				if !foundPath {
					t.Error("path label not found")
				}
			},
		},
		{
			name: "labels with CEL validations",
			cueContent: `
version: "1.0.0"
info: {
	title: "Test"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title: "Default Service"
			version: "1.0.0"
		}
		metrics: {
			api_requests: {
				namespace: "api"
				subsystem: "gateway"
				type: "counter"
				help: "API requests"
				labels: {
					method: {
						description: "HTTP method"
						validations: [
							"value in ['GET', 'POST', 'PUT', 'DELETE']"
						]
					}
				}
			}
		}
	}
}`,
			wantErr: false,
			checks: func(t *testing.T, spec *domain.Specification) {
				service := spec.Services["default"]
				metric := service.Metrics["api_requests"]

				if len(metric.Labels) != 1 {
					t.Fatalf("len(Labels) = %d, want 1", len(metric.Labels))
				}

				methodLabel := metric.Labels[0]
				if methodLabel.Name != "method" {
					t.Errorf("Label name = %q, want %q", methodLabel.Name, "method")
				}

				if len(methodLabel.Validations) != 1 {
					t.Errorf("len(Validations) = %d, want 1", len(methodLabel.Validations))
				}

				if methodLabel.Validations[0] != "value in ['GET', 'POST', 'PUT', 'DELETE']" {
					t.Errorf("Validation = %q, want enum validation", methodLabel.Validations[0])
				}
			},
		},
		{
			name: "missing required field - domain validation failure",
			cueContent: `
version: "1.0.0"
info: {
	version: "1.0.0"
}
services: {}`,
			wantErr: true,
		},
		{
			name: "histogram metric type",
			cueContent: `
version: "1.0.0"
info: {
	title: "Test"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title: "Default Service"
			version: "1.0.0"
		}
		metrics: {
			request_duration: {
				namespace: "http"
				subsystem: "server"
				type: "histogram"
				help: "Request duration"
				buckets: [0.1, 0.5, 1.0, 5.0]
			}
		}
	}
}`,
			wantErr: false,
			checks: func(t *testing.T, spec *domain.Specification) {
				service := spec.Services["default"]
				metric := service.Metrics["request_duration"]

				if metric.Type != domain.MetricTypeHistogram {
					t.Errorf("Metric type = %q, want %q", metric.Type, domain.MetricTypeHistogram)
				}
			},
		},
		{
			name: "summary metric type",
			cueContent: `
version: "1.0.0"
info: {
	title: "Test"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title: "Default Service"
			version: "1.0.0"
		}
		metrics: {
			response_size: {
				namespace: "http"
				subsystem: "server"
				type: "summary"
				help: "Response size"
			}
		}
	}
}`,
			wantErr: false,
			checks: func(t *testing.T, spec *domain.Specification) {
				service := spec.Services["default"]
				metric := service.Metrics["response_size"]

				if metric.Type != domain.MetricTypeSummary {
					t.Errorf("Metric type = %q, want %q", metric.Type, domain.MetricTypeSummary)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse CUE content
			ctx := cuecontext.New()
			value := ctx.CompileString(tt.cueContent)
			if value.Err() != nil {
				t.Fatalf("Failed to compile CUE: %v", value.Err())
			}

			// Extract specification
			extractor := NewCueExtractor()
			spec, err := extractor.Extract(value)

			if (err != nil) != tt.wantErr {
				t.Errorf("CueExtractor.Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if spec == nil {
				t.Fatal("Extract() returned nil specification")
			}

			// Run additional checks
			if tt.checks != nil {
				tt.checks(t, spec)
			}
		})
	}
}

func TestCueExtractor_ExtractFromRealFile(t *testing.T) {
	// Test with the actual with_cue_mod/metrics.cue file
	loader := NewCueLoader()
	value, result, err := loader.LoadAndValidate("../../testdata/with_cue_mod/metrics.cue")
	if err != nil {
		t.Fatalf("LoadAndValidate() error = %v", err)
	}

	if result.HasErrors() {
		t.Fatalf("metrics.cue validation failed: %+v", result)
	}

	extractor := NewCueExtractor()
	spec, err := extractor.Extract(value)

	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if spec == nil {
		t.Fatal("Extract() returned nil specification")
	}

	// Basic validation
	if spec.Info.Title == "" {
		t.Error("Extracted specification has empty title")
	}

	if len(spec.Services) == 0 {
		t.Error("Extracted specification has no services")
	}
}

func TestNewCueExtractor(t *testing.T) {
	extractor := NewCueExtractor()

	if extractor == nil {
		t.Fatal("NewCueExtractor() returned nil")
	}
}

func TestCueExtractor_InvalidCUEValue(t *testing.T) {
	ctx := cuecontext.New()

	// Create an invalid/incomplete CUE value
	value := ctx.CompileString("invalid: {}")

	extractor := NewCueExtractor()
	_, err := extractor.Extract(value)

	// Should fail domain validation since it's missing required fields
	if err == nil {
		t.Error("Extract() should fail on invalid specification, got nil error")
	}
}
