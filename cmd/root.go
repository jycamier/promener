package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jycamier/promener/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile         string
	rulesDirs       []string
	severityOnError string
	verbosity       int
	logFormat       string
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Get verbosity: flag takes precedence over config
		v := verbosity
		if v == 0 {
			v = viper.GetInt("logger.level")
		}

		// Get format: flag takes precedence over config
		format := logFormat
		if format == "" {
			format = viper.GetString("logger.format")
		}
		if format == "" {
			format = "text"
		}

		logging.Init(v, os.Stderr, format)

		logging.Debug("configuration loaded",
			"config_file", viper.ConfigFileUsed(),
			"verbosity", v,
			"log_format", format,
		)

		return nil
	},
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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .promener.yaml)")
	rootCmd.PersistentFlags().StringSliceVar(&rulesDirs, "rules", nil, "directories containing Rego rules for validation (repeatable)")
	rootCmd.PersistentFlags().StringVar(&severityOnError, "severity-on-error", "error", "minimum severity level to trigger exit 1 (error, warning, info)")

	// Logging flags
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "increase verbosity (-v=warn, -vv=info, -vvv=debug, -vvvv=trace)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "", "log output format: text or json (default: text)")

	viper.BindPFlag("rules", rootCmd.PersistentFlags().Lookup("rules"))
	viper.BindPFlag("severity_on_error", rootCmd.PersistentFlags().Lookup("severity-on-error"))
	viper.BindPFlag("logger.level", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("logger.format", rootCmd.PersistentFlags().Lookup("log-format"))
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory and parents up to HOME
		curr, _ := os.Getwd()
		home, _ := os.UserHomeDir()

		for {
			viper.AddConfigPath(curr)
			if curr == home || curr == filepath.Dir(curr) {
				break
			}
			curr = filepath.Dir(curr)
		}

		viper.SetConfigType("yaml")
		viper.SetConfigName(".promener")
	}

	viper.SetEnvPrefix("promener")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is acceptable - user may not have one
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but has errors (e.g., malformed YAML)
			// Note: Can't use logging.Error here because logging isn't initialized yet
			fmt.Fprintf(os.Stderr, "Warning: error reading config file: %v\n", err)
		}
	}
}
