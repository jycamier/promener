# Vet Command

The `vet` command validates CUE specifications without generating code. It's designed for:

- **Pre-flight checks**: Validate specs before code generation
- **CI/CD integration**: Fail builds on invalid specifications
- **Development feedback**: Catch errors during editing
- **Schema compliance**: Ensure specs match the expected format

## Basic Usage

```bash
# Validate a specification
promener vet metrics.cue

# Machine-readable output for CI/CD
promener vet metrics.cue --format json
```

## Command Syntax

```
promener vet <file> [flags]

Arguments:
  file              Path to CUE specification file (required)

Flags:
  --format string   Output format: "text" or "json" (default "text")
  -h, --help        Help for vet command
```

## What Gets Validated

The vet command performs comprehensive validation:

### 1. CUE Syntax Validation

Checks for valid CUE syntax:
- Correct package declaration
- Valid field definitions
- Proper use of braces, brackets, and quotes
- Correct type annotations

**Example error**:
```
✗ Validation failed

CUE validation errors:
  - metrics.http_requests_total.type: value must be one of "counter", "gauge", "histogram", "summary"
    at line 15, column 9
```

### 2. Schema Validation

Validates against the embedded CUE schema for your version:
- Required fields are present
- Field types match schema
- Enum values are valid
- Nested structures are correct

**Example**:
```cue
version: "1.0.0"  // Loads v1 schema
```

The schema defines:
- Valid metric types (`counter`, `gauge`, `histogram`, `summary`)
- Required fields (`namespace`, `type`, `help`)
- Optional fields (`subsystem`, `labels`, `constLabels`)
- Nested structure constraints

### 3. Domain Validation

Checks business rules beyond schema validation:
- Histogram metrics must have `buckets`
- Summary metrics must have `objectives`
- Label names must be valid Prometheus identifiers
- Namespace and subsystem names follow conventions
- Environment variable syntax is correct

### 4. CEL Expression Validation

Validates all label validation expressions:
- CEL syntax is correct
- Expressions compile successfully
- Expressions reference the `value` variable
- Expressions return boolean results

**Example validation**:
```cue
validations: [
    "value in ['GET', 'POST']",  // ✓ Valid
    "value.matches('^[0-9]+$')", // ✓ Valid
    "value > 100",                // ✗ Error: comparing string to int
]
```

## Output Formats

### Text Format (Default)

Human-readable output for terminal use:

**Success**:
```bash
$ promener vet metrics.cue
✓ Validation successful

Specification: My Application Metrics (v1.0.0)
Services: 1
Total metrics: 3
```

**Failure**:
```bash
$ promener vet metrics.cue
✗ Validation failed

CUE validation errors:
  - services.default.metrics.http_requests_total.type: value must be one of "counter", "gauge", "histogram", "summary"
    at line 15, column 9

  - services.default.metrics.request_duration.buckets: field required for histogram metrics
    at line 28, column 5

Domain validation errors:
  - metric "cache_size": gauge metrics cannot have buckets
  - label "Method": label names must be lowercase (use "method")
```

The text format includes:
- ✓/✗ symbols for quick status recognition
- Error messages with file locations
- Line and column numbers for CUE errors
- Categorized errors (CUE vs domain)

### JSON Format

Machine-readable output for CI/CD pipelines:

```bash
$ promener vet metrics.cue --format json
```

**Success**:
```json
{
  "valid": true,
  "errors": [],
  "specification": {
    "title": "My Application Metrics",
    "version": "1.0.0",
    "services": 1,
    "totalMetrics": 3
  }
}
```

**Failure**:
```json
{
  "valid": false,
  "errors": [
    {
      "type": "cue",
      "path": "services.default.metrics.http_requests_total.type",
      "message": "value must be one of \"counter\", \"gauge\", \"histogram\", \"summary\"",
      "line": 15,
      "column": 9
    },
    {
      "type": "domain",
      "path": "services.default.metrics.cache_size",
      "message": "gauge metrics cannot have buckets"
    }
  ]
}
```

JSON output includes:
- `valid`: Boolean indicating overall validation status
- `errors`: Array of validation errors with details
- `specification`: Metadata about the spec (on success)
- `type`: Error category (`cue` or `domain`)
- `path`: CUE path to the invalid field
- `line`/`column`: Source location (for CUE errors)

## CI/CD Integration

### GitHub Actions

```yaml
name: Validate Metrics

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install Promener
        run: go install github.com/jycamier/promener@latest

      - name: Validate metrics specification
        run: promener vet metrics.cue --format json
```

### GitLab CI

```yaml
validate-metrics:
  image: golang:1.21
  stage: validate
  script:
    - go install github.com/jycamier/promener@latest
    - promener vet metrics.cue --format json
  only:
    changes:
      - metrics.cue
```

### Jenkins

```groovy
pipeline {
    agent any
    stages {
        stage('Validate Metrics') {
            steps {
                sh 'go install github.com/jycamier/promener@latest'
                sh 'promener vet metrics.cue --format json'
            }
        }
    }
}
```

### Pre-commit Hook

Add validation to your pre-commit hook:

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check if metrics.cue has been modified
if git diff --cached --name-only | grep -q "metrics.cue"; then
    echo "Validating metrics.cue..."
    if ! promener vet metrics.cue; then
        echo "Error: metrics.cue validation failed"
        exit 1
    fi
fi
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

## Exit Codes

The vet command uses standard exit codes:

- **0**: Validation successful
- **1**: Validation failed (errors found)

Use exit codes in scripts:

```bash
if promener vet metrics.cue; then
    echo "Validation passed"
    promener generate go -i metrics.cue -o ./metrics
else
    echo "Validation failed - fix errors before generating"
    exit 1
fi
```

## Common Validation Errors

### Missing Required Fields

```
✗ services.default.metrics.http_requests_total.help: field required
```

**Solution**: Add the missing field:
```cue
http_requests_total: {
    namespace: "http"
    type:      "counter"
    help:      "Total HTTP requests"  // Add this
}
```

### Invalid Metric Type

```
✗ services.default.metrics.my_metric.type: value must be one of "counter", "gauge", "histogram", "summary"
```

**Solution**: Use a valid type:
```cue
my_metric: {
    type: "counter"  // Not "Counter" or "COUNTER"
}
```

### Missing Buckets for Histogram

```
✗ metric "request_duration": histogram metrics must define buckets
```

**Solution**: Add buckets array:
```cue
request_duration: {
    type: "histogram"
    buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
}
```

### Missing Objectives for Summary

```
✗ metric "request_size": summary metrics must define objectives
```

**Solution**: Add objectives map:
```cue
request_size: {
    type: "summary"
    objectives: {
        "0.5":  0.05
        "0.9":  0.01
        "0.99": 0.001
    }
}
```

### Invalid CEL Expression

```
✗ label "method" validation expression: compilation failed: undeclared reference to 'val' (in container '')
```

**Solution**: Use `value` variable:
```cue
validations: [
    "value in ['GET', 'POST']"  // Not "val"
]
```

### Invalid Environment Variable Syntax

```
✗ constant label "environment": invalid environment variable syntax: ${ENV
```

**Solution**: Use correct syntax:
```cue
constLabels: {
    environment: {
        value: "${ENVIRONMENT}"       // Required var
        // or
        value: "${ENVIRONMENT:prod}"  // With default
    }
}
```

### Label Description Missing

```
✗ services.default.metrics.http_requests_total.labels.method: description field required
```

**Solution**: Add description:
```cue
labels: {
    method: {
        description: "HTTP method"
    }
}
```

## Validation Workflow

Recommended workflow for developing metrics specifications:

1. **Write CUE specification**:
   ```bash
   vim metrics.cue
   ```

2. **Validate early and often**:
   ```bash
   promener vet metrics.cue
   ```

3. **Fix errors** based on validation output

4. **Generate code** after validation passes:
   ```bash
   promener generate go -i metrics.cue -o ./metrics
   ```

5. **Commit** both specification and generated code:
   ```bash
   git add metrics.cue metrics/
   git commit -m "Add HTTP metrics"
   ```

## Validation in Makefile

Integrate validation into your build process:

```makefile
.PHONY: validate-metrics
validate-metrics:
	@echo "Validating metrics specification..."
	@promener vet metrics.cue

.PHONY: generate-metrics
generate-metrics: validate-metrics
	@echo "Generating metrics code..."
	@promener generate go -i metrics.cue -o ./metrics

.PHONY: all
all: validate-metrics generate-metrics
```

Usage:
```bash
make validate-metrics  # Just validate
make generate-metrics  # Validate then generate
```

## Version-Based Schema Loading

The vet command uses the `version` field to load the appropriate embedded schema:

```cue
version: "1.0.0"  // Uses v1 schema
version: "2.0.0"  // Would use v2 schema (when available)
```

This allows:
- **Backward compatibility**: Old specs work with newer Promener versions
- **Forward compatibility**: New features in new schema versions
- **Clear migration**: Version bumps indicate breaking changes

Schema files are embedded in the binary at:
- `schema/v1/schema.cue` for version 1.x
- `schema/v2/schema.cue` for version 2.x (future)

## Validation Performance

Validation is fast:
- **Small specs** (< 50 metrics): < 100ms
- **Medium specs** (50-200 metrics): 100-500ms
- **Large specs** (200+ metrics): 500ms-2s

Validation time includes:
- CUE compilation
- Schema loading and unification
- CEL expression compilation
- Domain rule checking

For CI/CD, this overhead is negligible compared to build times.

## Debugging Validation Issues

### Enable Verbose Output

For detailed debugging, use CUE's built-in tools:

```bash
# Check CUE syntax
cue vet metrics.cue

# Format CUE file
cue fmt metrics.cue

# Evaluate CUE
cue eval metrics.cue
```

### Check Schema Match

Compare your spec against the schema:

```bash
# Export embedded schema (during development)
cue export schema/v1/schema.cue

# Check your spec structure
cue eval metrics.cue
```

### Isolate Problems

Test subsets of your specification:

```cue
// Comment out metrics to isolate issues
metrics: {
    // working_metric: { ... }

    problematic_metric: {
        // ...
    }
}
```

## Best Practices

1. **Validate frequently**: Run vet after each change during development
2. **Use CI/CD**: Automate validation in your pipeline
3. **JSON in CI**: Use `--format json` for machine-readable output
4. **Pre-commit hooks**: Catch errors before committing
5. **Version control specs**: Track specifications alongside code
6. **Document constraints**: Use comments to explain validation rules
7. **Test invalid specs**: Keep test files with known errors for regression testing

## See Also

- [CUE Specification](cue-specification.md) - Complete CUE format reference
- [Label Validation](label-validation.md) - CEL validation expressions
- [CUE Language](https://cuelang.org/) - Official CUE documentation
