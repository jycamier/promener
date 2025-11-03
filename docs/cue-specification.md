# CUE Specification Reference

Promener uses [CUE (Configure Unify Execute)](https://cuelang.org/) as the source format for metric specifications. CUE provides type safety, validation, and a powerful constraint system that ensures your metrics are correctly defined before code generation.

## Why CUE?

CUE offers several advantages over YAML:

- **Type safety**: CUE validates types at parse time, catching errors early
- **Schema validation**: Built-in schema validation with expressive constraints
- **Composability**: Import and reuse definitions across files
- **Expressiveness**: More concise syntax with less repetition
- **Tooling**: Rich ecosystem with formatting, validation, and IDE support

## Basic Structure

Every Promener specification follows this structure:

```cue
package main

version: "1.0.0"
info: {
    title:   "My Application Metrics"
    version: "1.0.0"
    package: "metrics"  // Optional: override generated package name
}
services: {
    default: {
        info: {
            title:   "Default Service"
            version: "1.0.0"
        }
        metrics: {
            // Metrics defined here
        }
    }
}
```

## Schema Version

The `version` field at the root level determines which CUE schema is used for validation. Promener embeds schemas for each major version and uses this field to select the appropriate one.

```cue
version: "1.0.0"  // Uses v1 schema
```

Current supported versions:
- **v1**: Initial schema with full feature support

## Top-Level Fields

### `version` (required)
- **Type**: `string`
- **Description**: Schema version for validation
- **Example**: `"1.0.0"`

### `info` (required)
- **Type**: `#Info` object
- **Description**: Metadata about the specification
- **Fields**:
  - `title` (required): Human-readable title
  - `description` (optional): Longer description
  - `version` (required): Specification version
  - `package` (optional): Override generated package name

### `services` (required)
- **Type**: Map of service definitions
- **Description**: One or more services containing metrics
- **Note**: Use `"default"` as the service name for single-service applications

## Service Definition

Each service contains:

```cue
services: {
    myservice: {
        info: {
            title:       "My Service"
            description: "Service description"
            version:     "1.0.0"
        }
        servers: [  // Optional
            {
                url:         "http://localhost:8080"
                description: "Local development server"
            },
        ]
        metrics: {
            // Metric definitions
        }
    }
}
```

## Metric Definition

### Basic Structure

```cue
metrics: {
    metric_name: {
        namespace:  "namespace"    // Required
        subsystem:  "subsystem"    // Optional
        type:       "counter"      // Required: counter, gauge, histogram, summary
        help:       "Description"  // Required
        labels: {                  // Optional
            // Label definitions
        }
        constLabels: {             // Optional
            // Constant label definitions
        }
        buckets: [...]             // Required for histograms
        objectives: {...}          // Required for summaries
        examples: {                // Optional
            // PromQL and alert examples
        }
    }
}
```

### Metric Types

Promener supports all Prometheus metric types:

#### Counter
```cue
http_requests_total: {
    namespace: "http"
    subsystem: "server"
    type:      "counter"
    help:      "Total HTTP requests"
}
```

#### Gauge
```cue
active_connections: {
    namespace: "http"
    subsystem: "server"
    type:      "gauge"
    help:      "Number of active connections"
}
```

#### Histogram
```cue
request_duration_seconds: {
    namespace: "http"
    subsystem: "server"
    type:      "histogram"
    help:      "HTTP request duration in seconds"
    buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
}
```

#### Summary
```cue
request_size_bytes: {
    namespace: "http"
    subsystem: "server"
    type:      "summary"
    help:      "HTTP request size in bytes"
    objectives: {
        "0.5":  0.05
        "0.9":  0.01
        "0.99": 0.001
    }
}
```

## Labels

Labels can be defined in two ways:

### Simple Labels (Description Only)

```cue
labels: {
    method: {
        description: "HTTP method"
    }
    status: {
        description: "HTTP status code"
    }
}
```

### Labels with Validation

Use CEL expressions to validate label values at runtime:

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

See [Label Validation](label-validation.md) for more details on CEL expressions.

## Constant Labels

Constant labels are static labels attached to all observations of a metric. They support environment variable substitution:

```cue
constLabels: {
    version: {
        value:       "1.0.0"
        description: "Application version"
    }
    environment: {
        value:       "${ENVIRONMENT}"  // Required env var
        description: "Deployment environment"
    }
    region: {
        value:       "${REGION:us-east-1}"  // Optional with default
        description: "AWS region"
    }
}
```

See [Constant Labels](constant-labels.md) for more details.

## Examples

Add PromQL queries and alert rules to your metrics for better documentation:

```cue
examples: {
    promql: [
        {
            query:       "rate(http_server_requests_total[5m])"
            description: "Request rate per second over 5 minutes"
        },
        {
            query:       "sum by (status) (http_server_requests_total)"
            description: "Total requests grouped by status"
        },
    ]
    alerts: [
        {
            name:        "HighErrorRate"
            expr:        "rate(http_server_requests_total{status=~\"5..\"}[5m]) > 0.1"
            description: "HTTP 5xx error rate is above 10%"
            for:         "5m"
            severity:    "critical"
            labels: {
                team: "backend"
            }
            annotations: {
                summary: "High error rate detected"
            }
        },
    ]
}
```

These examples appear in the generated HTML documentation.

## Deprecation

Mark metrics as deprecated to guide users toward replacements:

```cue
metrics: {
    old_metric: {
        namespace: "http"
        subsystem: "server"
        type:      "counter"
        help:      "Deprecated metric"
        deprecated: {
            since:      "2024-01-15"
            replacedBy: "new_metric_name"
            reason:     "Better accuracy with histogram"
        }
    }
}
```

See [Metric Deprecation](metric-deprecation.md) for more details.

## CUE Modules

Promener supports CUE modules, allowing you to organize and reuse metric definitions across multiple files:

### Project Structure

```
myproject/
├── cue.mod/
│   └── module.cue
├── metrics.cue
├── common_labels.cue
└── http_metrics.cue
```

### Module Definition

```cue
// cue.mod/module.cue
module: "github.com/myorg/myproject"
language: version: "v0.8.0"
```

### Importing Definitions

```cue
// common_labels.cue
package main

#CommonLabels: {
    environment: {
        value:       "${ENVIRONMENT}"
        description: "Deployment environment"
    }
    version: {
        value:       "${VERSION:1.0.0}"
        description: "Application version"
    }
}
```

```cue
// http_metrics.cue
package main

services: {
    default: {
        metrics: {
            http_requests_total: {
                namespace: "http"
                subsystem: "server"
                type:      "counter"
                help:      "Total HTTP requests"
                constLabels: #CommonLabels
            }
        }
    }
}
```

## Complete Example

Here's a comprehensive example showcasing all features:

```cue
package main

version: "1.0.0"
info: {
    title:       "E-commerce API Metrics"
    description: "Metrics for the e-commerce API service"
    version:     "2.1.0"
    package:     "metrics"
}

services: {
    default: {
        info: {
            title:       "E-commerce API"
            description: "Main API service for e-commerce platform"
            version:     "2.1.0"
        }
        servers: [
            {
                url:         "https://api.example.com"
                description: "Production API"
            },
        ]
        metrics: {
            http_requests_total: {
                namespace: "http"
                subsystem: "server"
                type:      "counter"
                help:      "Total number of HTTP requests"
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
                    endpoint: {
                        description: "API endpoint path"
                    }
                }
                constLabels: {
                    service: {
                        value:       "ecommerce-api"
                        description: "Service identifier"
                    }
                    environment: {
                        value:       "${ENVIRONMENT}"
                        description: "Deployment environment"
                    }
                }
                examples: {
                    promql: [
                        {
                            query:       "rate(http_server_requests_total[5m])"
                            description: "Request rate per second"
                        },
                        {
                            query:       "sum by (status) (http_server_requests_total)"
                            description: "Requests by status code"
                        },
                    ]
                    alerts: [
                        {
                            name:        "HighErrorRate"
                            expr:        "rate(http_server_requests_total{status=~\"5..\"}[5m]) > 0.1"
                            description: "HTTP 5xx error rate exceeds 10%"
                            for:         "5m"
                            severity:    "critical"
                        },
                    ]
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
                    service: {
                        value:       "ecommerce-api"
                        description: "Service identifier"
                    }
                }
                examples: {
                    promql: [
                        {
                            query:       "histogram_quantile(0.95, rate(http_server_request_duration_seconds_bucket[5m]))"
                            description: "95th percentile latency"
                        },
                    ]
                    alerts: [
                        {
                            name:        "HighLatency"
                            expr:        "histogram_quantile(0.95, rate(http_server_request_duration_seconds_bucket[5m])) > 1"
                            description: "95th percentile latency exceeds 1 second"
                            for:         "10m"
                            severity:    "warning"
                        },
                    ]
                }
            }
            active_connections: {
                namespace: "http"
                subsystem: "server"
                type:      "gauge"
                help:      "Number of active HTTP connections"
                constLabels: {
                    service: {
                        value:       "ecommerce-api"
                        description: "Service identifier"
                    }
                }
            }
        }
    }
}
```

## Validation

Before generating code, validate your CUE specification:

```bash
# Human-readable output
promener vet metrics.cue

# Machine-readable output for CI/CD
promener vet metrics.cue --format json
```

The vet command checks:
- CUE syntax and structure
- Schema compliance (version-based)
- Domain-level constraints
- Label validation expressions (CEL syntax)

See [Vet Command](vet-command.md) for more details.

## CUE Schema

Promener embeds CUE schemas for each major version. The current v1 schema defines:

```cue
#Info: {
    title:        string
    description?: string
    version:      string
    package?:     string
}

#Metric: {
    namespace:  string
    subsystem?: string
    type:       "counter" | "gauge" | "histogram" | "summary"
    help:       string
    labels?: [string]: {
        description: string
        validations?: [...string]
    }
    constLabels?: [string]: {
        value:       string
        description: string
    }
    buckets?: [...number]
    objectives?: [string]: number
    examples?: {
        promql?: [...#PromQLExample]
        alerts?: [...#AlertExample]
    }
}

#Promener: {
    version: string | *"1.0"
    info:    #Info
    services?: [string]: {
        info: #Info
        servers?: [...#Server]
        metrics: [string]: #Metric
    }
}
```

The schema is located in `schema/v1/schema.cue` and embedded in the binary at build time.

## Migration from YAML

If you have existing YAML specifications, convert them to CUE:

### YAML
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
```

### CUE Equivalent
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
    }
}
```

Key differences:
1. CUE uses `{}` for objects instead of indentation
2. CUE requires `"` for strings
3. Labels must have descriptions (enforced by schema)
4. CUE uses `:` for definitions and values

## Best Practices

1. **Use meaningful namespaces**: Group related metrics by domain (e.g., `http`, `db`, `cache`)
2. **Add subsystems**: Further organize metrics within namespaces (e.g., `http.server`, `http.client`)
3. **Document labels**: Always provide clear label descriptions
4. **Add validations**: Use CEL expressions to validate label values
5. **Include examples**: Add PromQL queries and alert rules for better documentation
6. **Use modules**: Split large specifications across multiple files
7. **Validate early**: Run `promener vet` before code generation
8. **Version your specs**: Track changes in version control alongside code

## See Also

- [Label Validation](label-validation.md) - CEL validation expressions
- [Vet Command](vet-command.md) - Specification validation
- [Constant Labels](constant-labels.md) - Environment variable substitution
- [Metric Deprecation](metric-deprecation.md) - Deprecating metrics
- [CUE Language](https://cuelang.org/) - Official CUE documentation
