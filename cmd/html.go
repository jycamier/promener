package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/jycamier/promener/internal/domain"
	"github.com/jycamier/promener/internal/signals"
	"github.com/jycamier/promener/internal/validator"
	"github.com/jycamier/promener/pkg/docs"
	"github.com/spf13/cobra"
)

var (
	htmlInputFiles []string
	htmlOutputFile string
	htmlWatch      time.Duration
)

// isURI returns true if the input string is a valid absolute URI
func isURI(input string) bool {
	u, err := url.Parse(input)
	return err == nil && u.IsAbs()
}

// loadSpecFromInput loads a spec from either a CUE file path or URI
func loadSpecFromInput(input string) (*domain.Specification, error) {
	v := validator.New()

	if isURI(input) {
		// Download CUE from URI to temporary file
		resp, err := http.Get(input)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch URI: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
		}

		// Create temp file for the downloaded CUE
		tmpFile, err := os.CreateTemp("", "promener_*.cue")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		// Write downloaded content
		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			return nil, fmt.Errorf("failed to write temp file: %w", err)
		}

		// Validate and extract from temp file
		spec, result, err := v.ValidateAndExtract(tmpFile.Name())
		if err != nil || result.HasErrors() {
			return nil, fmt.Errorf("validation failed for URI %s: %w", input, err)
		}
		return spec, nil
	}

	// Local file
	spec, result, err := v.ValidateAndExtract(input)
	if err != nil || result.HasErrors() {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return spec, nil
}

// htmlCmd represents the html command
var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "Generate HTML documentation from CUE specification",
	Long: `Generate beautiful HTML documentation for your Prometheus metrics.

The HTML documentation includes:
- Interactive search and filtering
- Dark mode support
- PromQL query examples with copy button
- Grafana dashboard examples
- Alertmanager alert rule examples
- Detailed label descriptions

Input sources can be local CUE files or URIs (http/https).

Examples:
  # Single file
  promener html -i metrics.cue -o docs/metrics.html

  # From URI
  promener html -i https://example.com/metrics.cue -o docs/metrics.html

  # Multiple files (aggregated into one HTML)
  promener html -i api.cue -i users.cue -i orders.cue -o docs/metrics.html

  # Mix of files and URIs
  promener html -i metrics.cue -i https://example.com/remote.cue -o docs/metrics.html

  # With watch mode
  promener html -i metrics.cue -o docs/metrics.html --watch 5s
  promener html -i api.cue -i users.cue -o docs/metrics.html --watch 5s`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(htmlInputFiles) == 0 {
			return fmt.Errorf("at least one input file is required")
		}

		generateHTML := func() error {
			// Load all specifications
			var builder docs.Builder

			if len(htmlInputFiles) == 1 {
				// Single file or URI: use simple generation
				spec, err := loadSpecFromInput(htmlInputFiles[0])
				if err != nil {
					return fmt.Errorf("failed to load spec: %w", err)
				}

				if err := docs.GenerateHTMLFile(spec, htmlOutputFile); err != nil {
					return fmt.Errorf("failed to generate HTML: %w", err)
				}
			} else {
				// todo: make a better / custom title and version
				builder = docs.NewHTMLBuilder("Aggregated Metrics", "1.0.0")

				for _, inputFile := range htmlInputFiles {
					spec, err := loadSpecFromInput(inputFile)
					if err != nil {
						return fmt.Errorf("failed to load spec %s: %w", inputFile, err)
					}
					builder.AddFromSpec(spec)
				}

				if err := builder.BuildHTMLFile(htmlOutputFile); err != nil {
					return fmt.Errorf("failed to generate HTML: %w", err)
				}
			}

			return nil
		}

		// Initial generation
		if err := generateHTML(); err != nil {
			return err
		}
		fmt.Printf("✓ Generated HTML documentation: %s\n", htmlOutputFile)

		// Watch mode
		if htmlWatch > 0 {
			fmt.Printf("👀 Watching for changes (every %s)... Press Ctrl+C to stop\n", htmlWatch)

			// Setup context with signal handling for graceful shutdown
			// Uses platform-specific signals (Unix: SIGINT+SIGTERM, Windows: only SIGINT)
			ctx, stop := signal.NotifyContext(context.Background(), signals.Shutdown()...)
			defer stop()

			ticker := time.NewTicker(htmlWatch)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					fmt.Printf("\n✓ Received shutdown signal, stopping watch mode...\n")
					return nil
				case <-ticker.C:
					if err := generateHTML(); err != nil {
						fmt.Printf("⚠ Error regenerating HTML: %v\n", err)
						continue
					}
					fmt.Printf("✓ Regenerated HTML documentation: %s (%s)\n", htmlOutputFile, time.Now().Format("15:04:05"))
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(htmlCmd)

	htmlCmd.Flags().StringSliceVarP(&htmlInputFiles, "input", "i", []string{}, "Input CUE specification (file path or URI) - can be specified multiple times (required)")
	htmlCmd.Flags().StringVarP(&htmlOutputFile, "output", "o", "", "Output HTML file (required)")
	htmlCmd.Flags().DurationVar(&htmlWatch, "watch", 0, "Watch for changes and regenerate (e.g., 5s, 1m)")

	htmlCmd.MarkFlagRequired("input")
	htmlCmd.MarkFlagRequired("output")
}
