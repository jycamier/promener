# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Promener is a code generator for Prometheus metrics that creates type-safe, organized Go code from YAML specifications. It generates structured metrics code organized by namespace/subsystem, with optional Uber FX dependency injection support and HTML documentation generation.

## Build and Development Commands

### Building
```bash
go build                           # Build the binary (outputs to ./promener)
go build -o promener              # Build with explicit output name
```

### Testing
```bash
go test ./...                      # Run all tests
go test ./internal/parser          # Run parser tests only
go test ./internal/generator       # Run generator tests only
go test ./internal/domain          # Run domain tests only
go test -v ./...                   # Run tests with verbose output
```

### Running
```bash
# Generate metrics code
./promener generate -i testdata/metrics.yaml -o output.go

# Generate with FX module
./promener generate -i testdata/metrics.yaml -o output.go --fx

# Generate HTML documentation
./promener html -i testdata/metrics.yaml -o docs/metrics.html

# Override package name
./promener generate -i metrics.yaml -o output.go -p mymetrics
```

### Development with go generate
```bash
go generate ./...                  # Run all go:generate directives
```

## Architecture

### Core Components

**Domain Models** (`internal/domain/`)
- `Specification`: Top-level structure representing the complete YAML spec (OpenAPI-inspired format)
- `Metric`: Individual metric definition with namespace, subsystem, type, labels, and optional constant labels
- `Labels`: Flexible label definitions supporting both simple string arrays and detailed maps with descriptions
- `ConstLabels`: Static labels with support for environment variable substitution (e.g., `${ENVIRONMENT:production}`)
- `MetricType`: Counter, Gauge, Histogram, Summary
- All domain types include validation methods

**Parser** (`internal/parser/`)
- Reads YAML files and unmarshals into `domain.Specification`
- Enriches metrics by populating the Name field from YAML map keys if not explicitly set
- Validates the specification after parsing
- Uses `gopkg.in/yaml.v3` for YAML parsing

**Generator** (`internal/generator/`)
- Transforms `domain.Specification` into Go code using text templates
- Organizes metrics hierarchically: `metrics.{Namespace}.{Subsystem}.Method()`
- Two main templates:
  - `registry_template.go`: Main metrics registry with type-safe methods
  - `fx_template.go`: Uber FX dependency injection module (optional via `--fx` flag)
- `model.go`: Builds template data by grouping metrics into namespaces/subsystems
- `envvar.go`: Handles environment variable substitution in constant labels
- Generates thread-safe initialization using `sync.Once`
- Uses `go/format` to format generated code

**HTML Generator** (`internal/htmlgen/`)
- Generates interactive HTML documentation from YAML specifications
- Features: search, dark mode, PromQL examples, Grafana/Alertmanager examples

**Commands** (`cmd/`)
- `root.go`: Base cobra command
- `generate.go`: Code generation command with flags for input, output, package override, and FX module
- `html.go`: HTML documentation generation command
- `version.go`: Version information

### Code Flow

1. **Input**: YAML file defining metrics with namespace, subsystem, type, labels, and optional constant labels
2. **Parse**: `parser.ParseFile()` → `domain.Specification` (validated)
3. **Transform**: `generator.buildTemplateData()` organizes metrics by namespace/subsystem into `TemplateData`
4. **Generate**: Templates execute to produce Go code with:
   - Metric collectors organized in nested structs
   - Type-safe methods with named parameters (one per label)
   - Environment variable resolution for constant labels
   - `sync.Once` for thread-safe initialization
5. **Format**: `go/format.Source()` formats the output
6. **Write**: Generated code written to file

### Naming Conventions

- **Snake case to CamelCase**: Metric names like `requests_total` become `RequestsTotal` for types/methods
- **Field names**: Use lowerCamelCase (e.g., `requestsTotal`)
- **Method names**: Use CamelCase with operation prefix:
  - Counter: `Inc{Name}()`, `Add{Name}(value)`
  - Gauge: `Set{Name}(value)`, `Inc{Name}()`, `Dec{Name}()`, `Add{Name}(value)`, `Sub{Name}(value)`
  - Histogram/Summary: `Observe{Name}(value)`
- **Namespaces/Subsystems**: CamelCase struct field names
- **Full metric names**: `namespace_subsystem_name` (underscore-separated)

### Generated Code Structure

```go
// For a metric with namespace=http, subsystem=server, name=requests_total
metrics := metrics.Default()
metrics.Http.Server.IncRequestsTotal("GET", "200", "/api")
```

Each subsystem gets:
- A struct with fields for each metric (e.g., `requestsTotal *prometheus.CounterVec`)
- Initialization methods that create collectors and register them
- Type-safe wrapper methods that accept label values as parameters

### Constant Labels and Environment Variables

Constant labels support environment variable substitution:
- `"${ENVIRONMENT}"`: Required env var (panics if missing)
- `"${ENVIRONMENT:production}"`: Optional with default value
- `"1.0.0"`: Static string value

The generator creates helper functions when needed:
- `os.Getenv()` for required vars
- `getEnvOrDefault()` for vars with defaults

### Testing

Test files follow the `_test.go` convention:
- `parser_test.go`: Tests YAML parsing and validation
- `metric_test.go`: Tests metric validation logic
- `envvar_test.go`: Tests environment variable parsing and substitution

Use `testdata/` directory for test fixtures (e.g., `testdata/metrics.yaml`)

## Key Design Decisions

1. **Hierarchical organization**: Metrics grouped by namespace and subsystem for logical structure
2. **Type safety**: Generated methods have typed parameters matching label names, preventing runtime errors
3. **Lazy initialization**: `sync.Once` ensures metrics are created only when first accessed
4. **Environment-aware**: Constant labels can reference environment variables with optional defaults
5. **Template-based generation**: Uses Go's `text/template` for flexible code generation
6. **FX integration**: Optional dependency injection support via `--fx` flag generates interfaces and providers
- c'est un générateur de code peu importe le language.