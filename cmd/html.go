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
	"github.com/jycamier/promener/internal/htmlgen"
	"github.com/jycamier/promener/internal/logging"
	"github.com/jycamier/promener/internal/signals"
	"github.com/jycamier/promener/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
func loadSpecFromInput(input string, rulesDirs []string) (*domain.Specification, error) {
	v := validator.New()
	if len(rulesDirs) > 0 {
		logging.Debug("loading rego rules", "directories", rulesDirs)
		v.SetRulesDirs(rulesDirs)
	}

	if isURI(input) {
		logging.Debug("fetching specification from URI", "uri", input)
		// Download CUE from URI to temporary file
		resp, err := http.Get(input)
		if err != nil {
			logging.Error("failed to fetch URI", "uri", input, "error", err)
			return nil, fmt.Errorf("failed to fetch URI: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logging.Error("HTTP request failed", "uri", input, "status", resp.StatusCode)
			return nil, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
		}

		// Create temp file for the downloaded CUE
		tmpFile, err := os.CreateTemp("", "promener_*.cue")
		if err != nil {
			logging.Error("failed to create temp file", "error", err)
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		logging.Debug("writing downloaded content to temp file", "temp_file", tmpFile.Name())
		// Write downloaded content
		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			logging.Error("failed to write temp file", "error", err)
			return nil, fmt.Errorf("failed to write temp file: %w", err)
		}

		// Validate and extract from temp file
		logging.Debug("validating specification from URI", "uri", input)
		spec, result, err := v.ValidateAndExtract(tmpFile.Name())
		if err != nil {
			logging.Error("validation failed for URI", "uri", input, "error", err)
			return nil, fmt.Errorf("validation failed for URI %s: %w", input, err)
		}
		if result.HasErrors() {
			errorCount := len(result.CueErrors) + len(result.DomainErrors) + len(result.RegoErrors)
			logging.Error("validation errors for URI", "uri", input, "error_count", errorCount)
			return nil, fmt.Errorf("validation failed for URI %s: specification contains %d errors", input, errorCount)
		}
		return spec, nil
	}

	// Local file
	logging.Debug("loading specification from file", "file", input)
	spec, result, err := v.ValidateAndExtract(input)
	if err != nil {
		logging.Error("validation failed", "file", input, "error", err)
		return nil, fmt.Errorf("validation failed for %s: %w", input, err)
	}
	if result.HasErrors() {
		errorCount := len(result.CueErrors) + len(result.DomainErrors) + len(result.RegoErrors)
		logging.Error("validation errors", "file", input, "error_count", errorCount)
		return nil, fmt.Errorf("validation failed for %s: specification contains %d errors", input, errorCount)
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
		inputFiles := viper.GetStringSlice("html.input")
		outputFile := viper.GetString("html.output")
		watch := viper.GetDuration("html.watch")
		rulesDirs := viper.GetStringSlice("rules")

		logging.Info("generating HTML documentation", "output", outputFile)
		logging.Debug("html options",
			"input_files", inputFiles,
			"watch", watch,
			"rules", rulesDirs,
		)

		if len(inputFiles) == 0 {
			return fmt.Errorf("at least one input file is required (via --input flag or config file)")
		}
		if outputFile == "" {
			return fmt.Errorf("output file is required (via --output flag or config file)")
		}

		generateHTML := func() error {
			// Load all specifications
			var builder *htmlgen.Builder

			if len(inputFiles) == 1 {
				// Single file or URI: use simple generation
				logging.Debug("loading single specification", "input", inputFiles[0])
				spec, err := loadSpecFromInput(inputFiles[0], rulesDirs)
				if err != nil {
					return fmt.Errorf("failed to load spec: %w", err)
				}

				logging.Debug("generating HTML from single spec")
				generator := htmlgen.NewGenerator()
				if err := generator.GenerateFile(spec, outputFile); err != nil {
					logging.Error("failed to generate HTML", "error", err)
					return fmt.Errorf("failed to generate HTML: %w", err)
				}
			} else {
				logging.Debug("aggregating multiple specifications", "count", len(inputFiles))
				builder = htmlgen.NewBuilder("Aggregated Metrics", "1.0.0")

				for _, inputFile := range inputFiles {
					logging.Debug("loading specification", "input", inputFile)
					spec, err := loadSpecFromInput(inputFile, rulesDirs)
					if err != nil {
						return fmt.Errorf("failed to load spec %s: %w", inputFile, err)
					}
					builder.AddFromSpec(spec)
				}

				logging.Debug("building aggregated HTML")
				if err := builder.Build(outputFile); err != nil {
					logging.Error("failed to generate HTML", "error", err)
					return fmt.Errorf("failed to generate HTML: %w", err)
				}
			}

			return nil
		}

		// Initial generation
		if err := generateHTML(); err != nil {
			return err
		}
		logging.Info("HTML documentation generated", "output", outputFile)
		fmt.Printf("✓ Generated HTML documentation: %s\n", outputFile)

		// Watch mode
		if watch > 0 {
			logging.Info("starting watch mode", "interval", watch)
			fmt.Printf("👀 Watching for changes (every %s)... Press Ctrl+C to stop\n", watch)

			// Setup context with signal handling for graceful shutdown
			// Uses platform-specific signals (Unix: SIGINT+SIGTERM, Windows: only SIGINT)
			ctx, stop := signal.NotifyContext(context.Background(), signals.Shutdown()...)
			defer stop()

			ticker := time.NewTicker(watch)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					logging.Info("received shutdown signal, stopping watch mode")
					fmt.Printf("\n✓ Received shutdown signal, stopping watch mode...\n")
					return nil
				case <-ticker.C:
					logging.Debug("regenerating HTML")
					if err := generateHTML(); err != nil {
						logging.Error("error regenerating HTML (will retry)", "error", err)
						fmt.Printf("⚠ Error regenerating HTML: %v\n", err)
						continue
					}
					logging.Debug("HTML regenerated successfully")
					fmt.Printf("✓ Regenerated HTML documentation: %s (%s)\n", outputFile, time.Now().Format("15:04:05"))
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(htmlCmd)

	htmlCmd.Flags().StringSliceVarP(&htmlInputFiles, "input", "i", []string{}, "Input CUE specification (file path or URI) - can be specified multiple times")
	htmlCmd.Flags().StringVarP(&htmlOutputFile, "output", "o", "", "Output HTML file")
	htmlCmd.Flags().DurationVar(&htmlWatch, "watch", 0, "Watch for changes and regenerate (e.g., 5s, 1m)")

	viper.BindPFlag("html.input", htmlCmd.Flags().Lookup("input"))
	viper.BindPFlag("html.output", htmlCmd.Flags().Lookup("output"))
	viper.BindPFlag("html.watch", htmlCmd.Flags().Lookup("watch"))
}
