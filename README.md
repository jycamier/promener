# Promener

**Promener** (from French "promener" - to walk through/explore) addresses a critical gap in observability: **the lack of structured, maintained documentation for Prometheus metrics**.

## The Problem

Most teams struggle with metric observability:
- Metrics are scattered across codebases without clear ownership or organization
- Documentation is absent, outdated, or disconnected from the actual implementation
- New team members can't easily understand what metrics exist or how to use them
- PromQL queries and alerting rules are tribal knowledge, not documented

## The Solution

Promener takes an **opinionated, Domain-Driven Design (DDD) approach** to metrics:

1. **Documentation-first**: Define metrics in a structured YAML specification where documentation is required, not optional
2. **Single source of truth**: Your YAML spec becomes the living documentation that's always in sync with code
3. **Domain organization**: Metrics are organized by namespace and subsystem, reflecting your business domains
4. **Generate everything**: From one spec, generate both production code AND beautiful, searchable HTML documentation

```yaml
metrics:
  requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total number of HTTP requests"  # Documentation required!
    labels:
      method:
        description: "HTTP method (GET, POST, etc.)"  # Label docs required!
    examples:
      promql:
        - query: 'rate(http_server_requests_total[5m])'
          description: "Request rate per second"
      alerts:
        - name: "HighErrorRate"
          expr: 'rate(http_server_requests_total{status=~"5.."}[5m]) > 0.1'
```

From this spec, Promener generates:
- **Type-safe code** (Go, .NET C#, Node.js TypeScript) with organized facades
- **Interactive HTML documentation** with searchable metrics, PromQL examples, and alert rules
- **Dependency injection modules** for easy integration
- **Thread-safe initialization** with proper registry management

**Supports Go**, **.NET (C#)**, and **Node.js (TypeScript)**.

## Features

- 📝 **YAML-based specifications** - Define metrics using an OpenAPI-inspired format
- 🌐 **Multi-language support** - Generate code for **Go**, **.NET (C#)**, and **Node.js (TypeScript)**
- 🏗️ **Organized structure** - Metrics grouped by namespace and subsystem
- 🔒 **Type-safe facades** - Generated methods with typed parameters
- 💉 **Dependency injection ready** - Supports Uber FX (Go) and Microsoft.Extensions.DependencyInjection (.NET)
- 🔄 **Thread-safe initialization** - Uses `sync.Once` for safe concurrent access
- 📊 **All metric types** - Counter, Gauge, Histogram, and Summary
- 🏷️ **Constant labels** - Support for static and environment variable-based labels
- ⚠️ **Metric deprecation** - Mark metrics as deprecated with migration guidance
- 🧪 **Mockable interfaces** - Generated interfaces for easy testing
- 📚 **Documentation generation** - Generate beautiful HTML documentation with examples
- 🔍 **Interactive docs** - Search, filter, dark mode, and copy-to-clipboard for queries

## Installation

### Option 1: Install as CLI tool

```bash
go install github.com/jycamier/promener@latest
```

Or build from source:

```bash
git clone https://github.com/jycamier/promener.git
cd promener
go build
```

### Option 2: Use with go:generate

Add to your `go.mod`:

```bash
go get github.com/jycamier/promener@latest
```

Then add a `//go:generate` directive to your code:

```go
//go:generate go run github.com/jycamier/promener generate go -i metrics.yaml -o metrics/ --di --fx
```

Or if you prefer using the installed tool:

```go
//go:generate promener generate go -i metrics.yaml -o metrics/ --di --fx
```

Then run:

```bash
go generate ./...
```

## Quick Start

### 1. Define your metrics in YAML

Create a `metrics.yaml` file:

```yaml
version: "1.0"
info:
  title: "My Application Metrics"
  version: "1.0.0"
  package: "metrics"

metrics:
  requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total number of HTTP requests"
    labels:
      - method
      - status
      - endpoint

  request_duration_seconds:
    namespace: http
    subsystem: server
    type: histogram
    help: "HTTP request duration in seconds"
    labels:
      - method
      - endpoint
    buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
    constLabels:
      environment: "${ENVIRONMENT:production}"
      region: "${REGION}"
      version: "1.0.0"
```

### 2. Generate code for your target language

#### For Go:

```bash
promener generate go -i metrics.yaml -o ./metrics
```

With Uber FX dependency injection:

```bash
promener generate go -i metrics.yaml -o ./metrics --di --fx
```

#### For .NET (C#):

```bash
promener generate dotnet -i metrics.yaml -o ./Metrics
```

With Dependency Injection extensions:

```bash
promener generate dotnet -i metrics.yaml -o ./Metrics --di
```

#### For Node.js (TypeScript):

```bash
promener generate nodejs -i metrics.yaml -o ./metrics
```

**Common options:**

- Override package/namespace name: `-p mymetrics`
- Generate DI code (Go): `--di --fx` (requires FX framework flag)
- Generate DI extensions (.NET): `--di`

### 3. Use in your application

#### Go Example

```go
package main

import (
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus/promhttp"
    "yourapp/metrics"
)

func main() {
    // Initialize metrics registry
    m := metrics.Default()

    // Expose metrics endpoint
    http.Handle("/metrics", promhttp.Handler())

    // Use metrics in your handlers
    http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Your handler logic
        w.Write([]byte("OK"))

        // Record metrics
        duration := time.Since(start).Seconds()
        m.Http.Server.IncRequestsTotal(r.Method, "200", "/api")
        m.Http.Server.ObserveRequestDurationSeconds(r.Method, "/api", duration)
    })

    http.ListenAndServe(":8080", nil)
}
```

#### .NET Example

```csharp
using Microsoft.AspNetCore.Builder;
using Microsoft.Extensions.DependencyInjection;
using Prometheus;
using YourNamespace.Metrics;

var builder = WebApplication.CreateBuilder(args);

// Register metrics with dependency injection
builder.Services.AddMetrics();

var app = builder.Build();

// Expose metrics endpoint
app.MapMetrics();

// Use metrics in your endpoints
app.MapGet("/api", (IHttpServerMetrics metrics) =>
{
    var stopwatch = System.Diagnostics.Stopwatch.StartNew();

    // Your endpoint logic
    var result = "OK";

    // Record metrics
    stopwatch.Stop();
    metrics.IncRequestsTotal("GET", "200", "/api");
    metrics.ObserveRequestDurationSeconds("GET", "/api", stopwatch.Elapsed.TotalSeconds);

    return result;
});

app.Run();
```

Alternatively, use the singleton pattern without DI:

```csharp
using Prometheus;
using YourNamespace.Metrics;

var app = WebApplication.CreateBuilder(args).Build();

// Get the default singleton instance
var metrics = MetricsRegistry.Default;

app.MapMetrics();

app.MapGet("/api", () =>
{
    metrics.HttpServer.IncRequestsTotal("GET", "200", "/api");
    return "OK";
});

app.Run();
```

## Documentation

- [HTTP Server Integration](docs/http-integration.md) - How to integrate metrics with HTTP servers
- [Constant Labels](docs/constant-labels.md) - Using static and environment-based constant labels
- [Metric Deprecation](docs/metric-deprecation.md) - How to deprecate metrics and guide migrations
- [YAML Specification](docs/yaml-specification.md) - Complete YAML format reference
- [Generated Code Structure](docs/generated-code.md) - Understanding the generated code

## Generated Code Structure

Promener generates metrics organized by namespace and subsystem:

```go
metrics := metrics.Default()

// Structure: metrics.{Namespace}.{Subsystem}.Method()
metrics.Http.Server.IncRequestsTotal("GET", "200", "/api")
metrics.Db.Postgres.ObserveQueryDurationSeconds("SELECT", "users", 0.002)
```

### Available Methods by Metric Type

**Counter:**
- `Inc{MetricName}(labels...)` - Increment by 1
- `Add{MetricName}(labels..., value)` - Add specific value

**Gauge:**
- `Set{MetricName}(labels..., value)` - Set to value
- `Inc{MetricName}(labels...)` - Increment by 1
- `Dec{MetricName}(labels...)` - Decrement by 1
- `Add{MetricName}(labels..., value)` - Add value
- `Sub{MetricName}(labels..., value)` - Subtract value

**Histogram:**
- `Observe{MetricName}(labels..., value)` - Observe a value

**Summary:**
- `Observe{MetricName}(labels..., value)` - Observe a value

### Constant Labels

Constant labels are automatically attached to all observations of a metric. They're useful for static metadata like environment, version, or region:

```yaml
metrics:
  requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total HTTP requests"
    labels:
      - method
      - status
    constLabels:
      version: "1.0.0"                      # Static value
      environment: "${ENVIRONMENT}"          # From env var
      region: "${REGION:eu-west-1}"         # With default value
```

The generated code automatically resolves environment variables at initialization:

```go
ConstLabels: prometheus.Labels{
    "version":     "1.0.0",
    "environment": os.Getenv("ENVIRONMENT"),
    "region":      getEnvOrDefault("REGION", "eu-west-1"),
}
```

### Dependency Injection with Uber FX

When using `--fx`, Promener generates an FX module with interfaces for easy testing:

```go
import (
    "go.uber.org/fx"
    "yourapp/metrics"
)

func main() {
    fx.New(
        metrics.Module,  // Provides *MetricsRegistry and all subsystem interfaces
        fx.Invoke(runServer),
    ).Run()
}

func runServer(httpMetrics metrics.HttpServerMetrics) {
    httpMetrics.IncRequestsTotal("GET", "200", "/api")
}
```

Each subsystem is provided as an interface for easy mocking in tests.

## Command Line Options

### Generate Command

The `generate` command now uses language-specific subcommands:

```
promener generate <language> [flags]

Subcommands:
  go        Generate Go code for Prometheus metrics
  dotnet    Generate .NET (C#) code for Prometheus metrics
  nodejs    Generate Node.js (TypeScript) code for Prometheus metrics

Global Flags:
  -i, --input string    Input YAML specification file (required)
  -o, --output string   Output directory (required)
```

#### Go Subcommand

```
promener generate go [flags]

Flags:
  -p, --package string  Override package name (optional)
  --di                  Generate dependency injection code (requires --fx)
  --fx                  Use Uber FX framework for DI
```

Examples:
```bash
# Generate Go code
promener generate go -i metrics.yaml -o ./metrics

# Generate Go code with Uber FX DI
promener generate go -i metrics.yaml -o ./metrics --di --fx

# Override package name
promener generate go -i metrics.yaml -o ./metrics -p mymetrics
```

#### .NET Subcommand

```
promener generate dotnet [flags]

Flags:
  -p, --package string  Override namespace (optional)
  --di                  Generate dependency injection extensions
```

Examples:
```bash
# Generate .NET code
promener generate dotnet -i metrics.yaml -o ./Metrics

# Generate .NET code with DI extensions
promener generate dotnet -i metrics.yaml -o ./Metrics --di

# Override namespace
promener generate dotnet -i metrics.yaml -o ./Metrics -p MyApp.Metrics
```

#### Node.js Subcommand

```
promener generate nodejs [flags]

Flags:
  -p, --package string  Override package name (optional)
```

Examples:
```bash
# Generate Node.js code
promener generate nodejs -i metrics.yaml -o ./metrics

# Override package name
promener generate nodejs -i metrics.yaml -o ./metrics -p my-metrics
```

### HTML Documentation Command

```
promener html [flags]

Flags:
  -i, --input string    Input YAML specification file (required)
  -o, --output string   Output HTML file (required)
```

## Documentation Generation

Promener can generate beautiful, interactive HTML documentation from your metrics specifications.

### Basic Documentation

```bash
promener html -i metrics.yaml -o docs/metrics.html
```

### Enhanced Documentation with Examples

For richer documentation, you can add descriptions, PromQL examples, alert rules, and deprecation warnings to your YAML:

```yaml
metrics:
  requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total number of HTTP requests"
    labels:
      method:
        description: "HTTP method (GET, POST, PUT, DELETE, etc.)"
      status:
        description: "HTTP status code (200, 404, 500, etc.)"
      endpoint:
        description: "API endpoint path"
    constLabels:
      environment:
        value: "${ENVIRONMENT:production}"
        description: "Deployment environment (production, staging, etc.)"
    deprecated:
      since: "2024-01-15"
      replacedBy: "request_duration_seconds"
      reason: "Switching to histogram for better latency tracking"
    examples:
      promql:
        - query: 'rate(http_server_requests_total[5m])'
          description: "Request rate per second over the last 5 minutes"
        - query: 'sum by (status) (http_server_requests_total)'
          description: "Total requests grouped by HTTP status code"
      alerts:
        - name: "HighErrorRate"
          expr: 'rate(http_server_requests_total{status=~"5.."}[5m]) > 0.1'
          description: "HTTP 5xx error rate is above 10%"
          for: "5m"
          severity: "critical"
```

### Generated Documentation Features

The HTML documentation includes:

- **Interactive search and filtering** - Quickly find metrics by name or namespace
- **Dark mode support** - Automatic theme switching
- **Deprecation warnings** - Visual warnings for deprecated metrics with migration info
- **Label descriptions** - Detailed information about each label
- **PromQL examples** - Ready-to-use queries with copy-to-clipboard
- **Alert rules** - Prometheus Alertmanager rule examples
- **Metric organization** - Grouped by namespace and subsystem with type badges
- **Full metric details** - Type, help text, labels, buckets/objectives

### Example Documentation

See [testdata/metrics_docs.html](testdata/metrics_docs.html) for a complete example generated from [testdata/metrics_with_docs.yaml](testdata/metrics_with_docs.yaml).

## Metric Deprecation

Promener supports marking metrics as deprecated with migration information. This helps teams manage metric lifecycle and guide users toward replacement metrics.

### Deprecation Syntax

```yaml
metrics:
  old_metric:
    namespace: http
    subsystem: server
    type: counter
    help: "Old metric - use new_metric instead"
    deprecated:
      since: "2024-01-15"           # When the metric was deprecated
      replacedBy: "new_metric_name"  # The replacement metric
      reason: "Switching to histogram for better latency tracking"
```

### Visual Indicators

In the generated HTML documentation:

- **Sidebar TOC**: Deprecated metrics show a ⚠️ warning icon next to their type badge
- **Metric Details**: A prominent orange warning banner displays:
  - When the metric was deprecated
  - Which metric replaces it
  - The reason for deprecation

This makes it easy for teams to identify deprecated metrics and understand migration paths.

## Examples

See the [testdata](testdata/) directory for complete examples.

## License

MIT

## TODO

- [ ] Standardize histogram buckets by business domain (e.g., HTTP latency, DB query duration, queue processing time)
- [ ] Standardize summary objectives by business domain (e.g., background jobs, batch processing, async tasks)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
