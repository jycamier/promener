package cmd

import (
	"fmt"
	"time"

	"github.com/jycamier/promener/pkg/docs"
	"github.com/spf13/cobra"
)

var (
	htmlInputFiles []string
	htmlOutputFile string
	htmlWatch      time.Duration
)

// htmlCmd represents the html command
var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "Generate HTML documentation from YAML specification",
	Long: `Generate beautiful HTML documentation for your Prometheus metrics.

The HTML documentation includes:
- Interactive search and filtering
- Dark mode support
- PromQL query examples with copy button
- Grafana dashboard examples
- Alertmanager alert rule examples
- Detailed label descriptions

Examples:
  # Single file
  promener html -i metrics.yaml -o docs/metrics.html

  # Multiple files (aggregated into one HTML)
  promener html -i api.yaml -i users.yaml -i orders.yaml -o docs/metrics.html

  # With watch mode
  promener html -i metrics.yaml -o docs/metrics.html --watch 5s
  promener html -i api.yaml -i users.yaml -o docs/metrics.html --watch 5s`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(htmlInputFiles) == 0 {
			return fmt.Errorf("at least one input file is required")
		}

		generateHTML := func() error {
			// Load all specifications
			var builder docs.Builder

			if len(htmlInputFiles) == 1 {
				// Single file: use simple generation
				spec, err := docs.LoadSpec(htmlInputFiles[0])
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
					spec, err := docs.LoadSpec(inputFile)
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
			ticker := time.NewTicker(htmlWatch)
			defer ticker.Stop()

			for range ticker.C {
				if err := generateHTML(); err != nil {
					fmt.Printf("⚠ Error regenerating HTML: %v\n", err)
					continue
				}
				fmt.Printf("✓ Regenerated HTML documentation: %s (%s)\n", htmlOutputFile, time.Now().Format("15:04:05"))
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(htmlCmd)

	htmlCmd.Flags().StringSliceVarP(&htmlInputFiles, "input", "i", []string{}, "Input YAML specification file(s) - can be specified multiple times (required)")
	htmlCmd.Flags().StringVarP(&htmlOutputFile, "output", "o", "", "Output HTML file (required)")
	htmlCmd.Flags().DurationVar(&htmlWatch, "watch", 0, "Watch for changes and regenerate (e.g., 5s, 1m)")

	htmlCmd.MarkFlagRequired("input")
	htmlCmd.MarkFlagRequired("output")
}
