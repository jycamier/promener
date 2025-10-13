package htmlgen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"

	"github.com/jycamier/promener/internal/domain"
)

//go:embed template.html
var htmlTemplate string

// MetricJSON represents a metric for JSON serialization
type MetricJSON struct {
	FullName    string          `json:"fullName"`
	Name        string          `json:"name"`
	Namespace   string          `json:"namespace"`
	Subsystem   string          `json:"subsystem"`
	Type        string          `json:"type"`
	Help        string          `json:"help"`
	Labels      []LabelJSON     `json:"labels"`
	ConstLabels []ConstLabelJSON `json:"constLabels,omitempty"`
	Examples    *ExamplesJSON   `json:"examples,omitempty"`
}

// LabelJSON represents a label for JSON serialization
type LabelJSON struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ConstLabelJSON represents a constant label for JSON serialization
type ConstLabelJSON struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// ExamplesJSON represents examples for JSON serialization
type ExamplesJSON struct {
	PromQL []PromQLExampleJSON `json:"promql,omitempty"`
	Alerts []AlertExampleJSON  `json:"alerts,omitempty"`
}

// PromQLExampleJSON represents a PromQL example for JSON
type PromQLExampleJSON struct {
	Query       string `json:"query"`
	Description string `json:"description,omitempty"`
}

// AlertExampleJSON represents an alert example for JSON
type AlertExampleJSON struct {
	Name        string `json:"name"`
	Expr        string `json:"expr"`
	Description string `json:"description,omitempty"`
	For         string `json:"for,omitempty"`
	Severity    string `json:"severity,omitempty"`
}

// TemplateData contains data for the HTML template
type TemplateData struct {
	Info        domain.Info
	MetricsJSON template.JS
}

// Generator handles HTML documentation generation
type Generator struct {
	tmpl *template.Template
}

// New creates a new HTML generator
func New() (*Generator, error) {
	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	return &Generator{tmpl: tmpl}, nil
}

// Generate generates HTML documentation from a specification
func (g *Generator) Generate(spec *domain.Specification) ([]byte, error) {
	// Convert metrics to JSON
	metrics := make([]MetricJSON, 0, len(spec.Metrics))
	for key, metric := range spec.Metrics {
		if metric.Name == "" {
			metric.Name = key
		}

		m := MetricJSON{
			FullName:  metric.FullName(),
			Name:      metric.Name,
			Namespace: metric.Namespace,
			Subsystem: metric.Subsystem,
			Type:      string(metric.Type),
			Help:      metric.Help,
		}

		// Convert labels
		for _, label := range metric.Labels {
			m.Labels = append(m.Labels, LabelJSON{
				Name:        label.Name,
				Description: label.Description,
			})
		}

		// Convert const labels
		for _, constLabel := range metric.ConstLabels {
			m.ConstLabels = append(m.ConstLabels, ConstLabelJSON{
				Name:        constLabel.Name,
				Value:       constLabel.Value,
				Description: constLabel.Description,
			})
		}

		// Convert examples if present
		if len(metric.Examples.PromQL) > 0 || len(metric.Examples.Alerts) > 0 {
			m.Examples = &ExamplesJSON{}

			for _, ex := range metric.Examples.PromQL {
				m.Examples.PromQL = append(m.Examples.PromQL, PromQLExampleJSON{
					Query:       ex.Query,
					Description: ex.Description,
				})
			}

			for _, ex := range metric.Examples.Alerts {
				m.Examples.Alerts = append(m.Examples.Alerts, AlertExampleJSON{
					Name:        ex.Name,
					Expr:        ex.Expr,
					Description: ex.Description,
					For:         ex.For,
					Severity:    ex.Severity,
				})
			}
		}

		metrics = append(metrics, m)
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metrics to JSON: %w", err)
	}

	// Prepare template data
	data := TemplateData{
		Info:        spec.Info,
		MetricsJSON: template.JS(jsonData),
	}

	// Execute template
	var buf []byte
	w := &bytesWriter{buf: &buf}
	if err := g.tmpl.Execute(w, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf, nil
}

// GenerateFile generates HTML and writes to a file
func (g *Generator) GenerateFile(spec *domain.Specification, outputPath string) error {
	html, err := g.Generate(spec)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, html, 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

// bytesWriter wraps a byte slice to implement io.Writer
type bytesWriter struct {
	buf *[]byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
