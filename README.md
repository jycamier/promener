# Promener

A code generator for Prometheus metrics that creates type-safe, organized Go code from YAML specifications.

## Features

- 📝 **YAML-based specifications** - Define metrics using an OpenAPI-inspired format
- 🏗️ **Organized structure** - Metrics grouped by namespace and subsystem
- 🔒 **Type-safe facades** - Generated methods with typed parameters
- 💉 **Dependency injection ready** - Supports custom Prometheus registerers with Uber FX
- 🔄 **Thread-safe initialization** - Uses `sync.Once` for safe concurrent access
- 📊 **All metric types** - Counter, Gauge, Histogram, and Summary
- 🏷️ **Constant labels** - Support for static and environment variable-based labels
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
//go:generate go run github.com/jycamier/promener generate -i metrics.yaml -o metrics/metrics_gen.go --fx
```

Or if you prefer using the installed tool:

```go
//go:generate promener generate -i metrics.yaml -o metrics/metrics_gen.go --fx
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

### 2. Generate Go code

**Option A: Using CLI directly**

```bash
promener generate -i metrics.yaml -o metrics/metrics.go
```

**Option B: Using go:generate with go run**

Create a file `metrics/doc.go`:

```go
package metrics

//go:generate go run github.com/jycamier/promener generate -i ../metrics.yaml -o metrics_gen.go --fx
```

Then run:

```bash
go generate ./metrics
```

**Option C: Using go:generate with installed tool**

Create a file `metrics/doc.go`:

```go
package metrics

//go:generate promener generate -i ../metrics.yaml -o metrics_gen.go --fx
```

Then run:

```bash
go generate ./metrics
```

**Additional options:**

- Override package name: `-p mymetrics`
- Generate FX module: `--fx`
- Example: `promener generate -i metrics.yaml -o metrics.go -p mymetrics --fx`

### 3. Use in your application

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

## Documentation

- [HTTP Server Integration](docs/http-integration.md) - How to integrate metrics with HTTP servers
- [Constant Labels](docs/constant-labels.md) - Using static and environment-based constant labels
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

```
promener generate [flags]

Flags:
  -i, --input string    Input YAML specification file (required)
  -o, --output string   Output Go file (required)
  -p, --package string  Override package name (optional)
  --fx                  Generate Uber FX module (optional)
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

For richer documentation, you can add descriptions, PromQL examples, and alert rules to your YAML:

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
- **Label descriptions** - Detailed information about each label
- **PromQL examples** - Ready-to-use queries with copy-to-clipboard
- **Alert rules** - Prometheus Alertmanager rule examples
- **Metric organization** - Grouped by namespace and subsystem
- **Full metric details** - Type, help text, labels, buckets/objectives

### Example Documentation

See [testdata/metrics_docs.html](testdata/metrics_docs.html) for a complete example generated from [testdata/metrics_with_docs.yaml](testdata/metrics_with_docs.yaml).

## Examples

See the [testdata](testdata/) directory for complete examples.

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
