# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Promener is a code generator for Prometheus metrics that creates type-safe, organized code from CUE specifications. It generates structured metrics code organized by namespace/subsystem, with CEL-based runtime label validation, optional Uber FX dependency injection support, and HTML documentation generation. CUE is used as the source of truth, providing both data definition and validation in one declarative language.

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
# Validate CUE specification (schema is embedded)
./promener vet testdata/test.cue

# Validate with JSON output for CI/CD
./promener vet testdata/test.cue --format json

# Generate Go metrics code
./promener generate go -i testdata/test.cue -o ./out

# Generate with FX dependency injection
./promener generate go -i testdata/test.cue -o ./out --di --fx

# Generate .NET code
./promener generate dotnet -i testdata/test.cue -o ./out

# Generate Node.js/TypeScript code
./promener generate nodejs -i testdata/test.cue -o ./out

# Generate HTML documentation
./promener html -i testdata/test.cue -o docs/metrics.html

# Override package name
./promener generate go -i metrics.cue -o ./out -p mymetrics
```

### Development with go generate
```bash
go generate ./...                  # Run all go:generate directives
```

## Example CUE Specification

A complete example with label validations:

```cue
package main

version: "1.0.0"
info: {
    title:   "API Metrics"
    version: "1.0.0"
}
services: {
    default: {
        info: {
            title:   "Default Service"
            version: "1.0.0"
        }
        metrics: {
            http_requests_total: {
                namespace: "http"
                subsystem: "server"
                type:      "counter"
                help:      "Total HTTP requests"
                labels: {
                    method: {
                        description: "HTTP method"
                        validations: [
                            "value in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']",
                        ]
                    }
                    status: {
                        description: "HTTP status code"
                        validations: [
                            "value.matches('^[1-5][0-9]{2}$')",
                        ]
                    }
                    service: {
                        description: "Service name (DNS-compatible)"
                        validations: [
                            "value.matches('^[a-z][a-z0-9-]*$')",
                            "size(value) >= 3",
                            "size(value) <= 63",
                        ]
                    }
                }
            }
        }
    }
}
```

This generates type-safe Go code with runtime validation:
```go
metrics := metrics.Default()
metrics.Http.Server.IncRequestsTotal("GET", "200", "api-gateway")
```

## Architecture

### Core Components

**Domain Models** (`internal/domain/`)
- `Specification`: Top-level structure representing the complete spec (OpenAPI-inspired format)
- `Metric`: Individual metric definition with namespace, subsystem, type, labels, and optional constant labels
- `Labels`: Flexible label definitions supporting both simple string arrays and detailed maps with descriptions and CEL validations
- `LabelDefinition`: Individual label with name, description, and optional CEL validation expressions
- `Validation`: CEL-based validation with compiled programs for runtime enforcement
- `ConstLabels`: Static labels with support for environment variable substitution (e.g., `${ENVIRONMENT:production}`)
- `MetricType`: Counter, Gauge, Histogram, Summary
- All domain types include validation methods

**Generator** (`internal/generator/`)
- Transforms `domain.Specification` into Go code using text templates
- Organizes metrics hierarchically: `metrics.{Namespace}.{Subsystem}.Method()`
- Two main templates:
  - `registry_template.go`: Main metrics registry with type-safe methods and CEL validation
  - `fx_template.go`: Uber FX dependency injection module (optional via `--fx` flag)
- `model.go`: Builds template data by grouping metrics into namespaces/subsystems with label definitions
- `envvar.go`: Handles environment variable substitution in constant labels
- Generates thread-safe initialization using `sync.Once`
- Compiles CEL validation expressions at initialization (one-time cost)
- Validates label values before each metric operation (panics on validation failure)
- Uses `go/format` to format generated code

**HTML Generator** (`internal/htmlgen/`)
- Generates interactive HTML documentation from YAML specifications
- Features: search, dark mode, PromQL examples, Grafana/Alertmanager examples

**Validator** (`internal/validator/`)
- `Validator`: Main validator that orchestrates CUE validation and extraction
- `CueLoader`: Loads CUE files, determines version, and validates against embedded schemas
- `CueExtractor`: Converts validated CUE values into `domain.Specification`
- `SchemaRegistry`: Maps major versions to embedded CUE schemas (v1, v2, etc.)
- Performs hybrid validation: CUE schema validation + domain rules
- `Formatter`: Formats validation results in text or JSON format
- Schemas are embedded in the binary at `schema/v1/`, `schema/v2/`, etc.
- Version is extracted from CUE's `version` field to load the appropriate schema

**Commands** (`cmd/`)
- `root.go`: Base cobra command
- `generate.go`: Code generation parent command (subcommands: go, dotnet, nodejs)
- `generate_go.go`: Go code generation with optional FX dependency injection
- `generate_dotnet.go`: .NET/C# code generation
- `generate_nodejs.go`: Node.js/TypeScript code generation
- `html.go`: HTML documentation generation command
- `vet.go`: Validation command with positional CUE file argument (schema embedded)
- `version.go`: Version information

### Code Flow

**Generation Flow:**
1. **Input**: CUE file defining metrics with namespace, subsystem, type, labels (with optional validations), and optional constant labels
2. **Validate**: `validator.ValidateAndExtract(cuePath)` → `domain.Specification`
   - Load CUE file and extract version
   - Load embedded schema for version (e.g., v1 from `schema/v1/`)
   - Validate CUE against schema
   - Extract into `domain.Specification` with label definitions and validations
   - Validate domain rules
3. **Transform**: `generator.buildTemplateData()` organizes metrics by namespace/subsystem into `TemplateData` with label validations
4. **Generate**: Templates execute to produce code with:
   - Metric collectors organized in nested structs
   - Type-safe methods with named parameters (one per label)
   - CEL validation expressions compiled at initialization
   - Runtime validation before each metric operation
   - Environment variable resolution for constant labels
   - `sync.Once` for thread-safe initialization (Go)
5. **Format**: Language-specific formatting (e.g., `go/format.Source()` for Go)
6. **Write**: Generated code written to output directory

**Validation Flow (vet command):**
1. **Input**: CUE file (positional argument)
2. **Load CUE**: `CueLoader.LoadAndValidate(cuePath)`
   - Read and compile CUE file
   - Extract `version` field
   - Load embedded schema via `SchemaRegistry.GetSchemaForVersion(version)`
3. **CUE Validation**:
   - Compile embedded schema
   - Unify specification with schema
   - Validate with `cue.Concrete(true)` and `cue.All()`
   - Collect validation errors with paths and line numbers
4. **Domain Validation** (if CUE passes):
   - Extract CUE to `domain.Specification`
   - Run `spec.Validate()` for domain-level checks
5. **Format Results**: Text (human-readable with ✓/✗) or JSON (machine-readable for CI/CD)
6. **Exit**: Code 0 (success) or 1 (validation errors found)

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

// Valid call - passes validation
metrics.Http.Server.IncRequestsTotal("GET", "200", "/api")

// Invalid call - panics with validation error
// metrics.Http.Server.IncRequestsTotal("INVALID", "200", "/api")
//   → panic: label "method" value "INVALID" failed validation
```

Each subsystem gets:
- A struct with fields for each metric (e.g., `requestsTotal *prometheus.CounterVec`)
- CEL validation programs compiled at initialization
- Initialization methods that create collectors and register them
- Type-safe wrapper methods that accept label values as parameters
- Automatic validation before each `WithLabelValues()` call

### Constant Labels and Environment Variables

Constant labels support environment variable substitution:
- `"${ENVIRONMENT}"`: Required env var (panics if missing)
- `"${ENVIRONMENT:production}"`: Optional with default value
- `"1.0.0"`: Static string value

The generator creates helper functions when needed:
- `os.Getenv()` for required vars
- `getEnvOrDefault()` for vars with defaults

### Label Validations with CEL

Labels support runtime validation using CEL (Common Expression Language):

**CUE Definition:**
```cue
labels: {
    method: {
        description: "HTTP method"
        validations: [
            "value in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']",
        ]
    }
    service: {
        description: "Service name"
        validations: [
            "value.matches('^[a-z][a-z0-9-]*$')",
            "size(value) >= 3",
            "size(value) <= 63",
        ]
    }
    status: {
        description: "HTTP status code"
        validations: [
            "value.matches('^[1-5][0-9]{2}$')",
        ]
    }
}
```

**Supported CEL Expressions:**
- **Enum validation**: `value in ['GET', 'POST', 'PUT']`
- **Regex matching**: `value.matches('^[a-z][a-z0-9-]*$')`
- **String operations**: `value.startsWith('prod-')`, `value.endsWith('-svc')`
- **Length constraints**: `size(value) >= 3 && size(value) <= 63`
- **Complex logic**: Combine multiple conditions with `&&`, `||`

**Generated Code Behavior:**
- CEL expressions compiled once at initialization (zero parsing overhead)
- Validations run before every `WithLabelValues()` call
- Validation failures trigger a panic with descriptive error message
- No validation overhead if label has no validation rules

**Example Usage:**
```go
metrics := metrics.Default()

// ✓ Valid - passes all validations
metrics.Http.Server.IncRequestsTotal("GET", "api-service", "200")

// ✗ Panics: label "method" value "INVALID" failed validation
metrics.Http.Server.IncRequestsTotal("INVALID", "api-service", "200")

// ✗ Panics: label "service" value "ab" failed validation (too short)
metrics.Http.Server.IncRequestsTotal("GET", "ab", "200")
```

**Dependencies:**
- `github.com/google/cel-go` v0.26+ for CEL evaluation

**Files:**
- `internal/domain/validation.go`: CEL parsing and validation logic
- `internal/domain/validation_test.go`: Comprehensive validation tests
- `internal/generator/templates/go/metrics.gotmpl`: Code generation with validation

### Testing

Test files follow the `_test.go` convention:
- `validator_test.go`: Tests CUE validation and extraction
- `schema_registry_test.go`: Tests version-based schema loading
- `formatter_test.go`: Tests validation result formatting (text/JSON)
- `metric_test.go`: Tests metric validation logic
- `validation_test.go`: Tests CEL expression parsing and runtime validation
- `envvar_test.go`: Tests environment variable parsing and substitution

Use `testdata/` directory for test fixtures:
- `testdata/test.cue`: Main CUE specification test file
- `testdata/test_validations.cue`: CUE file with CEL validation examples
- `testdata/with_cue_mod/`: CUE module tests with external imports
- `testdata/schemas/*.cue`: Additional CUE schema test files

## Key Design Decisions

1. **CUE as source of truth**: CUE provides both data definition and validation in one declarative language
2. **Embedded schemas**: CUE schemas are embedded in the binary, keyed by major version (v1, v2, etc.)
3. **Version-based validation**: The `version` field in CUE files determines which schema to use
4. **Hybrid validation**: CUE schema validation + domain-level validation for comprehensive checks
5. **Hierarchical organization**: Metrics grouped by namespace and subsystem for logical structure
6. **Type safety**: Generated methods have typed parameters matching label names, preventing runtime errors
7. **Lazy initialization**: `sync.Once` ensures metrics are created only when first accessed (Go)
8. **Environment-aware**: Constant labels can reference environment variables with optional defaults
9. **Template-based generation**: Uses language-specific templates for flexible code generation
10. **Multi-language support**: Go, .NET/C#, and Node.js/TypeScript generation from the same CUE spec
11. **FX integration**: Optional dependency injection support for Go via `--fx` flag
12. **CEL-based label validation**: Runtime validation using Google's CEL for expressive, safe validation rules with compile-time optimization