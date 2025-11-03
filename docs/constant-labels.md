# Constant Labels

Constant labels (ConstLabels) are labels that are automatically attached to every observation of a metric. Unlike dynamic labels that change with each observation, constant labels remain the same throughout the lifetime of your application.

## Table of Contents

- [When to Use Constant Labels](#when-to-use-constant-labels)
- [CUE Syntax](#cue-syntax)
- [Static Values](#static-values)
- [Environment Variables](#environment-variables)
- [Generated Code](#generated-code)
- [Best Practices](#best-practices)
- [Examples](#examples)

## When to Use Constant Labels

Constant labels are perfect for:

- **Application metadata**: version, build number, git commit
- **Deployment information**: environment (production, staging, dev)
- **Infrastructure details**: region, datacenter, availability zone
- **Instance identification**: hostname, pod name, container ID
- **Team/ownership**: team name, service owner

## CUE Syntax

Add `constLabels` to any metric definition:

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
                description: "Application version"
            }
            environment: {
                value:       "production"
                description: "Deployment environment"
            }
        }
    }
}
```

## Static Values

Use static values for information known at build time:

```cue
constLabels: {
    version: {
        value:       "1.0.0"
        description: "Application version"
    }
    build: {
        value:       "12345"
        description: "Build number"
    }
    service: {
        value:       "api-gateway"
        description: "Service name"
    }
}
```

Generated code:

```go
ConstLabels: prometheus.Labels{
    "version": "1.0.0",
    "build":   "12345",
    "service": "api-gateway",
}
```

## Environment Variables

### Basic Environment Variable

Use `${VAR_NAME}` to read from environment variables:

```cue
constLabels: {
    environment: {
        value:       "${ENVIRONMENT}"
        description: "Deployment environment"
    }
    region: {
        value:       "${AWS_REGION}"
        description: "AWS region"
    }
    hostname: {
        value:       "${HOSTNAME}"
        description: "Hostname"
    }
}
```

Generated code:

```go
ConstLabels: prometheus.Labels{
    "environment": os.Getenv("ENVIRONMENT"),
    "region":      os.Getenv("AWS_REGION"),
    "hostname":    os.Getenv("HOSTNAME"),
}
```

**Important**: If the environment variable is not set, the label value will be an empty string.

### Environment Variable with Default

Use `${VAR_NAME:default}` to provide a fallback value:

```cue
constLabels: {
    environment: {
        value:       "${ENVIRONMENT:production}"
        description: "Deployment environment"
    }
    region: {
        value:       "${AWS_REGION:us-east-1}"
        description: "AWS region"
    }
    datacenter: {
        value:       "${DATACENTER:dc1}"
        description: "Datacenter location"
    }
}
```

Generated code:

```go
ConstLabels: prometheus.Labels{
    "environment": getEnvOrDefault("ENVIRONMENT", "production"),
    "region":      getEnvOrDefault("AWS_REGION", "us-east-1"),
    "datacenter":  getEnvOrDefault("DATACENTER", "dc1"),
}

// Helper function is automatically generated
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## Generated Code

Promener automatically:

1. **Imports `os` package** when environment variables are used
2. **Generates helper function** `getEnvOrDefault` when defaults are specified
3. **Resolves values at initialization** time (when `NewMetricsRegistry()` is called)

### Example: Complete Generated Code

**CUE:**
```cue
metrics: {
    http_requests_total: {
        namespace: "http"
        subsystem: "server"
        type:      "counter"
        help:      "Total requests"
        labels: {
            method: {description: "HTTP method"}
        }
        constLabels: {
            version: {
                value:       "1.0.0"
                description: "Application version"
            }
            environment: {
                value:       "${ENVIRONMENT:production}"
                description: "Deployment environment"
            }
            region: {
                value:       "${REGION}"
                description: "AWS region"
            }
        }
    }
}
```

**Generated Go:**
```go
import (
    "os"
    "sync"
    "github.com/prometheus/client_golang/prometheus"
)

func NewMetricsRegistry(registerer prometheus.Registerer) *MetricsRegistry {
    once.Do(func() {
        registry = &MetricsRegistry{
            Http: &HttpMetrics{
                Server: &HttpServerMetricsImpl{
                    requestsTotal: prometheus.NewCounterVec(
                        prometheus.CounterOpts{
                            Namespace: "http",
                            Subsystem: "server",
                            Name:      "requests_total",
                            Help:      "Total requests",
                            ConstLabels: prometheus.Labels{
                                "version":     "1.0.0",
                                "environment": getEnvOrDefault("ENVIRONMENT", "production"),
                                "region":      os.Getenv("REGION"),
                            },
                        },
                        []string{"method"},
                    ),
                },
            },
        }
    })
    return registry
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## Best Practices

### 1. Keep Cardinality Low

Constant labels are included in every time series. Avoid high-cardinality values:

```cue
// ❌ BAD: High cardinality
constLabels: {
    request_id: {value: "${REQUEST_ID}"}  // Different for each request
    user_id: {value: "${USER_ID}"}        // Different for each user
}

// ✅ GOOD: Low cardinality
constLabels: {
    environment: {value: "${ENVIRONMENT:production}", description: "Environment"}
    region: {value: "${AWS_REGION:us-east-1}", description: "AWS region"}
    version: {value: "1.0.0", description: "Version"}
}
```

### 2. Use for Static Metadata Only

Constant labels can't change during runtime. Use dynamic labels for runtime values:

```cue
metrics: {
    http_requests_total: {
        type: "counter"
        labels: {
            method: {description: "HTTP method"}      // ✅ Dynamic: changes per request
            status: {description: "HTTP status"}      // ✅ Dynamic: changes per request
        }
        constLabels: {
            version: {value: "1.0.0", description: "Version"}  // ✅ Static: same for all requests
        }
    }
}
```

### 3. Provide Defaults for Safety

Always provide defaults for environment-based labels to prevent empty values:

```cue
// ❌ Risky: might be empty
constLabels: {
    environment: {value: "${ENVIRONMENT}"}
}

// ✅ Safe: has fallback
constLabels: {
    environment: {value: "${ENVIRONMENT:production}", description: "Environment"}
}
```

### 4. Use Meaningful Names

Choose label names that are clear and follow Prometheus conventions:

```cue
// ✅ GOOD
constLabels: {
    environment: {value: "${ENV:production}", description: "Environment"}
    region: {value: "${AWS_REGION:us-east-1}", description: "AWS region"}
    version: {value: "1.0.0", description: "Application version"}
}

// ❌ BAD
constLabels: {
    env: {value: "${E}"}
    reg: {value: "${R}"}
    v: {value: "1"}
}
```

### 5. Be Consistent Across Metrics

Use the same constant labels across related metrics:

```cue
// Global const labels that apply to all metrics
metrics: {
    http_requests_total: {
        // ... metric config ...
        constLabels: {
            environment: {value: "${ENVIRONMENT:production}", description: "Environment"}
            version: {value: "1.0.0", description: "Version"}
        }
    }
    http_request_duration_seconds: {
        // ... metric config ...
        constLabels: {
            environment: {value: "${ENVIRONMENT:production}", description: "Environment"}
            version: {value: "1.0.0", description: "Version"}
        }
    }
}
```

## Examples

### Example 1: Kubernetes Deployment

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
            pod: {value: "${POD_NAME}", description: "Pod name"}
            namespace: {value: "${POD_NAMESPACE}", description: "K8s namespace"}
            cluster: {value: "${CLUSTER_NAME:production}", description: "Cluster name"}
            version: {value: "${APP_VERSION:1.0.0}", description: "App version"}
        }
    }
}
```

### Example 2: Multi-Region Service

```cue
metrics: {
    api_calls_total: {
        namespace: "api"
        subsystem: "gateway"
        type:      "counter"
        help:      "Total API calls"
        labels: {
            endpoint: {description: "API endpoint"}
            method: {description: "HTTP method"}
        }
        constLabels: {
            region: {value: "${AWS_REGION:us-east-1}", description: "AWS region"}
            availability_zone: {value: "${AWS_AZ:us-east-1a}", description: "AZ"}
            environment: {value: "${ENVIRONMENT:production}", description: "Environment"}
        }
    }
}
```

### Example 3: Microservice with Version Tracking

```cue
metrics: {
    service_requests_total: {
        namespace: "myservice"
        subsystem: "api"
        type:      "counter"
        help:      "Total service requests"
        labels: {
            operation: {description: "Operation name"}
        }
        constLabels: {
            service: {value: "user-service", description: "Service name"}
            version: {value: "${BUILD_VERSION:dev}", description: "Build version"}
            commit: {value: "${GIT_COMMIT:unknown}", description: "Git commit"}
            build_date: {value: "${BUILD_DATE:unknown}", description: "Build date"}
        }
    }
}
```

### Example 4: Multi-Tenant Application

```cue
metrics: {
    queries_total: {
        namespace: "db"
        subsystem: "postgres"
        type:      "counter"
        help:      "Total database queries"
        labels: {
            operation: {description: "SQL operation"}
            table: {description: "Table name"}
        }
        constLabels: {
            tenant: {value: "${TENANT_ID}", description: "Tenant identifier"}
            environment: {value: "${ENVIRONMENT:production}", description: "Environment"}
            database: {value: "${DB_NAME:main}", description: "Database name"}
        }
    }
}
```

## Runtime Behavior

Constant labels are resolved **once** when the metrics registry is initialized:

```go
// Set environment variables before initialization
os.Setenv("ENVIRONMENT", "staging")
os.Setenv("REGION", "eu-west-1")

// Initialize registry - const labels are set now
registry := metrics.NewMetricsRegistry(prometheus.DefaultRegisterer)

// Changing env vars after initialization has NO effect
os.Setenv("ENVIRONMENT", "production")  // Too late!

// Metrics still have "staging" and "eu-west-1"
registry.Http.Server.IncRequestsTotal("GET", "200", "/api")
```

## Troubleshooting

### Empty Label Values

**Problem**: Label appears empty in Prometheus

**Solution**: Check environment variable is set before initialization:

```go
// Verify before creating registry
if os.Getenv("ENVIRONMENT") == "" {
    log.Fatal("ENVIRONMENT variable not set")
}

registry := metrics.NewMetricsRegistry(prometheus.DefaultRegisterer)
```

### Label Not Appearing

**Problem**: Constant label doesn't show in metrics

**Cause**: CUE syntax error or missing field

**Solution**: Check CUE formatting and ensure it's under `constLabels`:

```cue
constLabels: {
    environment: {value: "production", description: "Environment"}  // Correct structure
}
```

### High Cardinality Warning

**Problem**: Prometheus warns about high cardinality

**Cause**: Constant label has too many unique values

**Solution**: Use dynamic labels instead, or reduce label value variety

## Related Documentation

- [CUE Specification](cue-specification.md) - Complete CUE format reference
- [Label Validation](label-validation.md) - CEL validation expressions
- [HTTP Server Integration](http-integration.md) - Integration examples
