
<div style="display: flex; justify-content: center;">
  <img src="docs/img/logo.png" alt="small" width="300"/>
</div>

**Promener** (from French "promener" - to walk through/explore) addresses a critical gap in observability: **the lack of structured, maintained documentation for Prometheus metrics**.

<img src="docs/img/promener.png" alt="small" width="300"/>


> [!WARNING]
> This project is **experimental**. APIs, behavior, and structure may change without notice.


## The Problem

Most teams struggle with metric observability:
- Metrics are scattered across codebases without clear ownership or organization
- Documentation is absent, outdated, or disconnected from the actual implementation
- New team members can't easily understand what metrics exist or how to use them
- PromQL queries and alerting rules are tribal knowledge, not documented

## The Solution

Promener takes an **opinionated, Domain-Driven Design (DDD) approach** to metrics:

1. **Documentation-first**: Define metrics in a structured CUE specification where documentation is required, not optional
2. **Single source of truth**: Your CUE spec becomes the living documentation that's always in sync with code
3. **Domain organization**: Metrics are organized by namespace and subsystem, reflecting your business domains
4. **Generate everything**: From one spec, generate both production code AND beautiful, searchable HTML documentation

```cue
package main

version: "1.0.0"
info: {
    title:   "My Application Metrics"
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
                            "value in ['GET', 'POST', 'PUT', 'DELETE']",
                        ]
                    }
                }
            }
        }
    }
}
```

From this spec, Promener generates:
- **Type-safe code** (Go, .NET C#, Node.js TypeScript) with organized facades
- **Runtime label validation** using CEL (Common Expression Language)
- **Interactive HTML documentation** with searchable metrics, PromQL examples, and alert rules
- **Dependency injection modules** for easy integration
- **Thread-safe initialization** with proper registry management

**Supports Go**, **.NET (C#)**, and **Node.js (TypeScript)**.

## Features

- 📝 **CUE-based specifications** - Define metrics using CUE language with built-in validation
- ✅ **Schema validation** - Embedded CUE schemas validate your specifications before generation
- 🔍 **Vet command** - Validate metrics specifications without generating code
- 🛡️ **Label validation (currently only supported for **GO** code generate)** - Label validation using CEL (Common Expression Language)
- 🌐 **Multi-language support** - Generate code for **Go**, **.NET (C#)**, and **Node.js (TypeScript)**
- 🏗️ **Organized structure** - Metrics grouped by namespace and subsystem
- 🔒 **Type-safe facades** - Generated methods with typed parameters
- 💉 **Dependency injection ready** - Supports Uber FX (Go) and Microsoft.Extensions.DependencyInjection (.NET)
- 📊 **All metric types** - Counter, Gauge, Histogram, and Summary
- 🏷️ **Constant labels** - Support for static and environment variable-based labels
- ⚠️ **Metric deprecation** - Mark metrics as deprecated with migration guidance
- 🧪 **Mockable interfaces** - Generated interfaces for easy testing
- 📚 **Documentation generation** - Generate beautiful HTML documentation with examples
- 🔍 **Interactive docs** - Search, filter, dark mode, and copy-to-clipboard for queries
- 📦 **CUE module support** - Use CUE modules with external imports

## Installation

### With Homebrew

```bash
brew tap jycamier/homebrew-tap
brew install promener
```

### With Go

```bash
go install github.com/jycamier/promener@latest
```

### Use with go:generate

Add to your `go.mod`:

```bash
go get -tool github.com/jycamier/promener@latest
```

Then add a `//go:generate` directive to your code:

```go
//go:generate go tool github.com/jycamier/promener generate go --input metrics.cue --ouput metrics/ --di --fx
```

Or if you prefer to run the generator directly:

```go
//go:generate go run github.com/jycamier/promener@latest generate go --input metrics.cue --ouput metrics/ --di --fx
```

Then run:

```bash
go generate ./...
```


## Quick Start

### 1. Define your metrics in CUE

Create a `metrics.cue` file:

```cue
package main

version: "1.0.0"
info: {
    title:   "My Application Metrics"
    version: "1.0.0"
    package: "metrics"
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
                help:      "Total number of HTTP requests"
                labels: {
                	application: {
                		description: "The application name"
                		inherited:   "Injected via Prometheus relabeling from the pod label `app`"
                	}
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
                }
            }
            http_request_duration_seconds: {
                namespace: "http"
                subsystem: "server"
                type:      "histogram"
                help:      "HTTP request duration in seconds"
                labels: {
                    method: {
                        description: "HTTP method"
                    }
                    endpoint: {
                        description: "API endpoint"
                    }
                }
                buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
                constLabels: {
                    environment: {
                        value:       "${ENVIRONMENT:production}"
                        description: "Deployment environment"
                    }
                    version: {
                        value:       "1.0.0"
                        description: "Application version"
                    }
                }
            }
        }
    }
}
```

### 1.1 Validate your specification

Before generating code, validate your CUE specification:

```bash
promener vet metrics.cue
```

This validates your specification against the embedded CUE schema and checks for errors. Use `--format json` for machine-readable output in CI/CD pipelines.

### 2. Generate code for your target language

#### For Go:

```bash
promener generate go -i metrics.cue -o ./metrics
```

With Uber FX dependency injection:

```bash
promener generate go -i metrics.cue -o ./metrics --di --fx
```

#### For .NET (C#):

```bash
promener generate dotnet -i metrics.cue -o ./Metrics
```

With Dependency Injection extensions:

```bash
promener generate dotnet -i metrics.cue -o ./Metrics --di
```

#### For Node.js (TypeScript):

```bash
promener generate nodejs -i metrics.cue -o ./metrics
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

- [CUE Specification](docs/cue-specification.md) - Complete CUE format reference and schema documentation
- [Rego Validation](docs/rego-validation.md) - Define and enforce validation policies with Rego
- [Golden Signals](docs/golden-signals.md) - Define and document the four key SRE signals (Latency, Errors, Traffic, Saturation)
- [Label Validation](docs/label-validation.md) - Using CEL for runtime label validation
- [Vet Command](docs/vet-command.md) - Validating specifications before code generation
- [HTTP Server Integration](docs/http-integration.md) - How to integrate metrics with HTTP servers
- [Constant Labels](docs/constant-labels.md) - Using static and environment-based constant labels
- [Metric Deprecation](docs/metric-deprecation.md) - How to deprecate metrics and guide migrations
- [GitHub Pages](docs/github-pages.md) - Live documentation example and deployment guide

### Live Example

See a live example of generated HTML documentation at: **https://jycamier.github.io/promener/**

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

```cue
metrics: {
    http_requests_total: {
        namespace: "http"
        subsystem: "server"
        type:      "counter"
        help:      "Total HTTP requests"
        labels: {
            method: {description: "HTTP method"}
            status: {description: "HTTP status"}
        }
        constLabels: {
            version: {
                value:       "1.0.0"
                description: "Static version"
            }
            environment: {
                value:       "${ENVIRONMENT}"
                description: "From env var"
            }
            region: {
                value:       "${REGION:eu-west-1}"
                description: "With default value"
            }
        }
    }
}
```

The generated code automatically resolves environment variables at initialization:

```go
ConstLabels: prometheus.Labels{
    "version":     "1.0.0",
    "environment": os.Getenv("ENVIRONMENT"),
    "region":      getEnvOrDefault("REGION", "eu-west-1"),
}
```

### Inherited Labels

Inherited labels are labels **owned and managed by infrastructure**, not by your application code. These labels are injected by infrastructure components such as:
- Prometheus relabeling rules
- Service mesh sidecars (Istio, Linkerd)
- Kubernetes operators or admission controllers
- Cloud provider metadata services
- Observability platforms (Datadog, Grafana Agent)

Common examples include:
- `cluster`: Kubernetes cluster name
- `region`, `zone`: Cloud provider location
- `pod`, `namespace`, `node`: Kubernetes metadata
- `instance`: Service instance identifier
- `env`, `datacenter`: Infrastructure-level environment identifiers

#### Why Document Inherited Labels?

While your application doesn't set these labels, documenting them is crucial because:
1. **Clear ownership**: Makes it explicit which labels are application vs infrastructure responsibility
2. **Query awareness**: Developers need to know all available labels for PromQL queries
3. **Cardinality planning**: Inherited labels affect metric cardinality and storage costs
4. **Documentation completeness**: Generated docs show the complete label set for each metric

#### Defining Inherited Labels

Use the `inherited` field to document the infrastructure mechanism responsible for the label:

```cue
metrics: {
    http_requests_total: {
        namespace: "http"
        subsystem: "server"
        type:      "counter"
        help:      "Total HTTP requests"
        labels: {
            // Application labels (set in code, validated at runtime)
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
            // Infrastructure labels (injected by infrastructure, no validation)
            cluster: {
                description: "Kubernetes cluster name"
                inherited:   "Injected via Prometheus relabeling from kubernetes_sd_config metadata"
            }
            region: {
                description: "Cloud provider region"
                inherited:   "Added by infrastructure relabeling using EC2 instance metadata"
            }
            mesh_version: {
                description: "Service mesh version"
                inherited:   "Injected by Istio sidecar proxy"
            }
        }
    }
}
```

**Note**: Inherited labels should NOT have `validations` since they are controlled by infrastructure, not your application.

#### Generated Code Behavior

**Application labels** (without `inherited`) are included in generated method signatures:
```go
// Only application labels (method, status) appear as parameters
metrics.Http.Server.IncRequestsTotal("GET", "200")
// Infrastructure labels (cluster, region, mesh_version) are added by infrastructure
```

**Inherited labels** are excluded from method signatures but remain part of the metric definition for documentation and query purposes.

#### HTML Documentation

In the generated HTML documentation, labels are organized into two groups:
- **Application**: Labels your code sets directly (your responsibility)
- **Infrastructure**: Inherited labels with explanations of their injection mechanism (infrastructure team responsibility)

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

## Configuration File

Promener supports a `.promener.yaml` configuration file to store your command-line options. This avoids repeating flags for every command.

The tool searches for `.promener.yaml` in the current directory and recursively up to your `$HOME` directory.

Example `.promener.yaml`:

```yaml
input: metrics.cue
output: ./gen/metrics

go:
  package: metrics
  di: true
  fx: true

dotnet:
  package: MyCompany.Metrics
  di: true

nodejs:
  package: my-metrics

html:
  output: docs/metrics.html
  watch: 5s
```

When a configuration file is present, you can simply run commands without arguments:

```bash
promener generate go
promener html
```

CLI flags always take precedence over configuration file settings.

## Command Line Options

### Vet Command

Validate CUE specifications without generating code:

```
promener vet <file> [flags]

Flags:
  --format string   Output format: text or json (default "text")

Examples:
  promener vet metrics.cue                    # Human-readable output
  promener vet metrics.cue --format json      # Machine-readable for CI/CD
```

### Generate Command

The `generate` command now uses language-specific subcommands:

```
promener generate <language> [flags]

Subcommands:
  go        Generate Go code for Prometheus metrics
  dotnet    Generate .NET (C#) code for Prometheus metrics
  nodejs    Generate Node.js (TypeScript) code for Prometheus metrics

Global Flags:
  -i, --input string    Input CUE specification file (required)
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
promener generate go -i metrics.cue -o ./metrics

# Generate Go code with Uber FX DI
promener generate go -i metrics.cue -o ./metrics --di --fx

# Override package name
promener generate go -i metrics.cue -o ./metrics -p mymetrics
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
promener generate dotnet -i metrics.cue -o ./Metrics

# Generate .NET code with DI extensions
promener generate dotnet -i metrics.cue -o ./Metrics --di

# Override namespace
promener generate dotnet -i metrics.cue -o ./Metrics -p MyApp.Metrics
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
promener generate nodejs -i metrics.cue -o ./metrics

# Override package name
promener generate nodejs -i metrics.cue -o ./metrics -p my-metrics
```

### HTML Documentation Command

```
promener html [flags]

Flags:
  -i, --input string    Input CUE specification file (required)
  -o, --output string   Output HTML file (required)
```

## Documentation Generation

Promener can generate beautiful, interactive HTML documentation from your metrics specifications.

### Basic Documentation

```bash
promener html -i metrics.cue -o docs/metrics.html
```

### Enhanced Documentation with Examples

For richer documentation, you can add descriptions, PromQL examples, alert rules, and deprecation warnings to your CUE specification:

```cue
metrics: {
    http_requests_total: {
        namespace: "http"
        subsystem: "server"
        type:      "counter"
        help:      "Total number of HTTP requests"
        labels: {
            method: {
                description: "HTTP method (GET, POST, PUT, DELETE, etc.)"
            }
            status: {
                description: "HTTP status code (200, 404, 500, etc.)"
            }
            endpoint: {
                description: "API endpoint path"
            }
        }
        constLabels: {
            environment: {
                value:       "${ENVIRONMENT:production}"
                description: "Deployment environment (production, staging, etc.)"
            }
        }
        examples: {
            promql: [
                {
                    query:       "rate(http_server_requests_total[5m])"
                    description: "Request rate per second over the last 5 minutes"
                },
                {
                    query:       "sum by (status) (http_server_requests_total)"
                    description: "Total requests grouped by HTTP status code"
                },
            ]
            alerts: [
                {
                    name:        "HighErrorRate"
                    expr:        "rate(http_server_requests_total{status=~\"5..\"}[5m]) > 0.1"
                    description: "HTTP 5xx error rate is above 10%"
                    for:         "5m"
                    severity:    "critical"
                },
            ]
        }
    }
}
```

### Generated Documentation Example

TODO

## Label Validation with CEL

Promener supports runtime label validation using CEL (Common Expression Language). This ensures that label values meet your requirements before being recorded.

```cue
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
```

The generated code validates labels at runtime, panicking on validation failures with descriptive error messages. See the [Label Validation](docs/label-validation.md) documentation for more details.

## Metric Deprecation

Promener supports marking metrics as deprecated with migration information. This helps teams manage metric lifecycle and guide users toward replacement metrics.

### Deprecation Syntax

```cue
metrics: {
    old_metric: {
        namespace: "http"
        subsystem: "server"
        type:      "counter"
        help:      "Old metric - use new_metric instead"
        deprecated: {
            since:      "2024-01-15"
            replacedBy: "new_metric_name"
            reason:     "Switching to histogram for better latency tracking"
        }
    }
}
```

## Examples

See the [example](testdata/with_cue_mod) directory for complete examples.

## TODO

- [x] Standardize histogram buckets by business domain (e.g., HTTP latency, DB query duration, queue processing time)
- [x] Standardize summary objectives by business domain (e.g., background jobs, batch processing, async tasks)
- [x] Standardize metrics for common usacase
- [x] Config file for promener
- [x] A way to contribute standard / common metrics

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
