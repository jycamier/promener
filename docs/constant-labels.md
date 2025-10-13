# Constant Labels

Constant labels (ConstLabels) are labels that are automatically attached to every observation of a metric. Unlike dynamic labels that change with each observation, constant labels remain the same throughout the lifetime of your application.

## Table of Contents

- [When to Use Constant Labels](#when-to-use-constant-labels)
- [YAML Syntax](#yaml-syntax)
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

## YAML Syntax

Add `constLabels` to any metric definition:

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
      version: "1.0.0"
      environment: "production"
```

## Static Values

Use static values for information known at build time:

```yaml
constLabels:
  version: "1.0.0"
  build: "12345"
  service: "api-gateway"
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

```yaml
constLabels:
  environment: "${ENVIRONMENT}"
  region: "${AWS_REGION}"
  hostname: "${HOSTNAME}"
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

```yaml
constLabels:
  environment: "${ENVIRONMENT:production}"
  region: "${AWS_REGION:us-east-1}"
  datacenter: "${DATACENTER:dc1}"
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

**YAML:**
```yaml
metrics:
  requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total requests"
    labels:
      - method
    constLabels:
      version: "1.0.0"
      environment: "${ENVIRONMENT:production}"
      region: "${REGION}"
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

```yaml
# ❌ BAD: High cardinality
constLabels:
  request_id: "${REQUEST_ID}"  # Different for each request
  user_id: "${USER_ID}"        # Different for each user

# ✅ GOOD: Low cardinality
constLabels:
  environment: "${ENVIRONMENT:production}"
  region: "${AWS_REGION:us-east-1}"
  version: "1.0.0"
```

### 2. Use for Static Metadata Only

Constant labels can't change during runtime. Use dynamic labels for runtime values:

```yaml
metrics:
  requests_total:
    type: counter
    labels:
      - method      # ✅ Dynamic: changes per request
      - status      # ✅ Dynamic: changes per request
    constLabels:
      version: "1.0.0"  # ✅ Static: same for all requests
```

### 3. Provide Defaults for Safety

Always provide defaults for environment-based labels to prevent empty values:

```yaml
# ❌ Risky: might be empty
constLabels:
  environment: "${ENVIRONMENT}"

# ✅ Safe: has fallback
constLabels:
  environment: "${ENVIRONMENT:production}"
```

### 4. Use Meaningful Names

Choose label names that are clear and follow Prometheus conventions:

```yaml
# ✅ GOOD
constLabels:
  environment: "${ENV:production}"
  region: "${AWS_REGION:us-east-1}"
  version: "1.0.0"

# ❌ BAD
constLabels:
  env: "${E}"
  reg: "${R}"
  v: "1"
```

### 5. Be Consistent Across Metrics

Use the same constant labels across related metrics:

```yaml
# Global const labels that apply to all metrics
metrics:
  requests_total:
    # ... metric config ...
    constLabels:
      environment: "${ENVIRONMENT:production}"
      version: "1.0.0"

  request_duration_seconds:
    # ... metric config ...
    constLabels:
      environment: "${ENVIRONMENT:production}"
      version: "1.0.0"
```

## Examples

### Example 1: Kubernetes Deployment

```yaml
metrics:
  http_requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total HTTP requests"
    labels:
      - method
      - status
    constLabels:
      pod: "${POD_NAME}"
      namespace: "${POD_NAMESPACE}"
      cluster: "${CLUSTER_NAME:production}"
      version: "${APP_VERSION:1.0.0}"
```

### Example 2: Multi-Region Service

```yaml
metrics:
  api_calls_total:
    namespace: api
    subsystem: gateway
    type: counter
    help: "Total API calls"
    labels:
      - endpoint
      - method
    constLabels:
      region: "${AWS_REGION:us-east-1}"
      availability_zone: "${AWS_AZ:us-east-1a}"
      environment: "${ENVIRONMENT:production}"
```

### Example 3: Microservice with Version Tracking

```yaml
metrics:
  service_requests_total:
    namespace: myservice
    subsystem: api
    type: counter
    help: "Total service requests"
    labels:
      - operation
    constLabels:
      service: "user-service"
      version: "${BUILD_VERSION:dev}"
      commit: "${GIT_COMMIT:unknown}"
      build_date: "${BUILD_DATE:unknown}"
```

### Example 4: Multi-Tenant Application

```yaml
metrics:
  queries_total:
    namespace: db
    subsystem: postgres
    type: counter
    help: "Total database queries"
    labels:
      - operation
      - table
    constLabels:
      tenant: "${TENANT_ID}"
      environment: "${ENVIRONMENT:production}"
      database: "${DB_NAME:main}"
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

**Cause**: YAML syntax error or missing tag

**Solution**: Check YAML formatting and ensure it's under `constLabels`:

```yaml
constLabels:
  environment: "production"  # Correct indentation
```

### High Cardinality Warning

**Problem**: Prometheus warns about high cardinality

**Cause**: Constant label has too many unique values

**Solution**: Use dynamic labels instead, or reduce label value variety

## Related Documentation

- [YAML Specification](yaml-specification.md)
- [Generated Code Structure](generated-code.md)
- [HTTP Server Integration](http-integration.md)
