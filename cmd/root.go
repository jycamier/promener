package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "promener",
	Short: "Generates structured Prometheus metrics code with named parameters and interactive HTML documentation from YAML specifications",
	Long: `Promener is a code generator for Prometheus metrics that creates type-safe,
organized code from YAML specifications.

Features:
- Type-safe metrics code organized by namespace and subsystem
- Generated methods with typed parameters (one per label)
- Support for all Prometheus metric types (Counter, Gauge, Histogram, Summary)
- Constant labels with environment variable substitution
- Optional dependency injection module generation (e.g., Uber FX)
- Thread-safe initialization
- Interactive HTML documentation with search, examples, and dark mode

Example workflow:
1. Define metrics in YAML with namespace, subsystem, type, and labels
2. Generate code: promener generate -i metrics.yaml -o metrics.{ext}
3. Use in your application with a clean, structured API`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.promener.yaml)")
}
