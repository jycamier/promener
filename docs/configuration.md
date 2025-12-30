# Configuration

Promener supports configuration through multiple sources with the following priority (highest to lowest):

1. **Command-line flags** - Override everything
2. **Environment variables** - Prefix: `PROMENER_`
3. **Configuration file** - `.promener.yaml`
4. **Default values**

## Configuration File

Promener looks for `.promener.yaml` in the current directory and parent directories up to your home directory.

### File Location

```bash
# Explicit path
promener --config /path/to/.promener.yaml vet metrics.cue

# Auto-discovery (searches current dir and parents)
promener vet metrics.cue
```

### Complete Example

```yaml
# .promener.yaml

# Input/output defaults
input: metrics.cue
output: ./generated

# Logging configuration
logger:
  level: 2      # 0=error, 1=warn, 2=info, 3=debug, 4=trace
  format: text  # text or json

# Validation settings
rules:
  - ./rules
  - ./policies
severity_on_error: error  # error, warning, or info

# Go generator settings
go:
  package: metrics
  di: false
  fx: false

# .NET generator settings
dotnet:
  package: Metrics
  di: false

# Node.js generator settings
nodejs:
  package: metrics

# HTML generator settings
html:
  input:
    - metrics.cue
    - api-metrics.cue
  output: docs/metrics.html
  watch: 0s

# Vet command settings
vet:
  format: text  # text or json
```

## Logging Configuration

Control log output verbosity and format.

### Verbosity Levels

| Level | Flag | Config | Description |
|-------|------|--------|-------------|
| ERROR | (none) | `0` | Errors only |
| WARN | `-v` | `1` | Warnings and errors |
| INFO | `-vv` | `2` | Informational messages |
| DEBUG | `-vvv` | `3` | Detailed debug output |
| TRACE | `-vvvv` | `4` | Very verbose tracing |

### Configuration

```yaml
logger:
  level: 2      # INFO level
  format: text  # Human-readable output
```

### Command-Line Override

```bash
# Override config file with flags
promener vet metrics.cue -vvv              # DEBUG level
promener vet metrics.cue --log-format=json # JSON format
```

### Output Examples

**Text format (default):**
```
time=2024-01-15T10:30:00.000+01:00 level=INFO msg="validating specification" file=metrics.cue
time=2024-01-15T10:30:00.005+01:00 level=INFO msg="validation successful"
```

**JSON format:**
```json
{"time":"2024-01-15T10:30:00.000+01:00","level":"INFO","msg":"validating specification","file":"metrics.cue"}
{"time":"2024-01-15T10:30:00.005+01:00","level":"INFO","msg":"validation successful"}
```

### Use Cases

| Use Case | Recommended Settings |
|----------|---------------------|
| Production CI/CD | `level: 0`, `format: json` |
| Development | `level: 2`, `format: text` |
| Debugging issues | `level: 3`, `format: text` |
| Verbose troubleshooting | `level: 4`, `format: text` |

## Environment Variables

All configuration options can be set via environment variables with the `PROMENER_` prefix.

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `PROMENER_LOGGER_LEVEL` | Log verbosity (0=error, 1=warn, 2=info, 3=debug, 4=trace) | `0` | `PROMENER_LOGGER_LEVEL=2` |
| `PROMENER_LOGGER_FORMAT` | Log output format (text, json) | `text` | `PROMENER_LOGGER_FORMAT=json` |
| `PROMENER_CACHE_DIR` | Cache directory for remote rules (Git, HTTP). **Env only.** | `~/.promener/cache` | `PROMENER_CACHE_DIR=/tmp/cache` |
| `PROMENER_INPUT` | Default input file path | (none) | `PROMENER_INPUT=metrics.cue` |
| `PROMENER_OUTPUT` | Default output directory | (none) | `PROMENER_OUTPUT=./generated` |
| `PROMENER_RULES` | Rego rule directories (comma-separated) | (none) | `PROMENER_RULES=./rules,./policies` |
| `PROMENER_SEVERITY_ON_ERROR` | Minimum severity to trigger exit 1 (error, warning, info) | `error` | `PROMENER_SEVERITY_ON_ERROR=warning` |
| `PROMENER_GO_PACKAGE` | Go package name override | (output dir name) | `PROMENER_GO_PACKAGE=mymetrics` |
| `PROMENER_GO_DI` | Generate Go DI code | `false` | `PROMENER_GO_DI=true` |
| `PROMENER_GO_FX` | Use Uber FX framework | `false` | `PROMENER_GO_FX=true` |
| `PROMENER_DOTNET_PACKAGE` | .NET namespace override | (output dir name) | `PROMENER_DOTNET_PACKAGE=MyApp.Metrics` |
| `PROMENER_DOTNET_DI` | Generate .NET DI extensions | `false` | `PROMENER_DOTNET_DI=true` |
| `PROMENER_NODEJS_PACKAGE` | Node.js package name override | (output dir name) | `PROMENER_NODEJS_PACKAGE=app-metrics` |
| `PROMENER_VET_FORMAT` | Vet command output format (text, json) | `text` | `PROMENER_VET_FORMAT=json` |
| `PROMENER_HTML_OUTPUT` | HTML output file path | (none) | `PROMENER_HTML_OUTPUT=docs/metrics.html` |
| `PROMENER_HTML_WATCH` | HTML watch interval | `0s` | `PROMENER_HTML_WATCH=5s` |

### Naming Convention

Config keys map to environment variables:
- Prefix: `PROMENER_`
- Dots (`.`) become underscores (`_`)
- Hyphens (`-`) become underscores (`_`)
- All uppercase

Examples: `logger.level` → `PROMENER_LOGGER_LEVEL`, `severity-on-error` → `PROMENER_SEVERITY_ON_ERROR`

## Global Flags

These flags are available for all commands:

```
Global Flags:
      --config string              Config file path (default: .promener.yaml)
      --log-format string          Log format: text or json (default: text)
      --rules strings              Rego rule directories (repeatable)
      --severity-on-error string   Exit 1 threshold: error, warning, info (default: error)
  -v, --verbose count              Increase verbosity (-v, -vv, -vvv, -vvvv)
```

## Command-Specific Configuration

### Generate Command

```yaml
# Common generate settings
input: metrics.cue
output: ./generated

# Go-specific
go:
  package: metrics    # Override package name
  di: true           # Generate DI code
  fx: true           # Use Uber FX framework

# .NET-specific
dotnet:
  package: MyApp.Metrics
  di: true

# Node.js-specific
nodejs:
  package: app-metrics
```

### Vet Command

```yaml
vet:
  format: json  # Output format for CI/CD
```

### HTML Command

```yaml
html:
  input:
    - api-metrics.cue
    - db-metrics.cue
  output: docs/metrics.html
  watch: 5s  # Auto-regenerate every 5 seconds
```

## Project-Specific Configuration

Create a `.promener.yaml` in your project root:

```yaml
# Project defaults
input: specs/metrics.cue
output: internal/metrics

logger:
  level: 1  # Show warnings in development

go:
  package: metrics
  di: true
  fx: true

rules:
  - ./validation-rules
```

## CI/CD Configuration

For CI/CD pipelines, use environment variables or a dedicated config:

```yaml
# .promener.ci.yaml
logger:
  level: 0      # Errors only (quiet)
  format: json  # Machine-readable

severity_on_error: warning  # Fail on warnings too

vet:
  format: json
```

```bash
promener --config .promener.ci.yaml vet metrics.cue
```

## Configuration Precedence Examples

### Example 1: Flag Overrides Config

```yaml
# .promener.yaml
logger:
  level: 2
  format: json
```

```bash
# Flag wins: uses DEBUG level and text format
promener vet metrics.cue -vvv --log-format=text
```

### Example 2: Environment Overrides Config

```yaml
# .promener.yaml
logger:
  level: 1
```

```bash
# Environment wins: uses INFO level
export PROMENER_LOGGER_LEVEL=2
promener vet metrics.cue
```

### Example 3: Mixed Sources

```yaml
# .promener.yaml
input: default.cue
logger:
  level: 1
```

```bash
# Result: input from flag, logger.level from config
promener vet custom.cue
# Uses: input=custom.cue, logger.level=1
```

## See Also

- [Vet Command](vet-command.md) - Validation command reference
- [Rego Validation](rego-validation.md) - Custom policy rules
- [CUE Specification](cue-specification.md) - Input file format
