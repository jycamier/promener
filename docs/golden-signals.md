# Golden Signals

Golden Signals are the four key metrics that Google SRE uses to monitor distributed systems. Promener allows you to define Golden Signals in your CUE specification to document operational health indicators and generate pre-computed recording rules.

## What are Golden Signals?

The four Golden Signals, as defined in the [Google SRE Book](https://sre.google/sre-book/monitoring-distributed-systems/), are:

| Signal | Description | Typical Metrics |
|--------|-------------|-----------------|
| **Latency** | How long it takes to service a request | Histograms (p50, p95, p99) |
| **Errors** | Rate of requests that fail | Counter ratios (5xx / total) |
| **Traffic** | Demand on your system | Request rate (req/s) |
| **Saturation** | How "full" your service is | Gauges (connections, memory, CPU) |

## Basic Structure

Golden Signals are defined per service, organized by topic (typically `namespace/subsystem`):

```cue
services: {
    myservice: {
        info: { ... }
        metrics: { ... }

        goldenSignals: {
            "http/server": {
                latency: { ... }
                errors: { ... }
                traffic: { ... }
                saturation: { ... }
            }
            "db/postgres": {
                latency: { ... }
                errors: { ... }
                traffic: { ... }
                saturation: { ... }
            }
        }
    }
}
```

## Topic Naming

Topics use the `namespace/subsystem` format to match your metrics organization:

- `http/server` - HTTP server metrics
- `http/client` - HTTP client (outgoing) metrics
- `db/postgres` - PostgreSQL database metrics
- `db/redis` - Redis database metrics
- `cache/redis` - Redis cache metrics
- `queue/rabbitmq` - RabbitMQ queue metrics

This allows you to define different Golden Signals for different parts of your system.

## Golden Signal Definition

Each signal has the following structure:

```cue
latency: {
    // Required: What this signal measures
    description: "HTTP request latency distribution"

    // Required: References to metrics defined in this service
    metrics: ["request_duration_seconds"]

    // Optional: Pre-computed queries for dashboards
    recordingRules: [
        {
            name:  "http:server:latency:p95:5m"
            query: "histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
        },
    ]

    // Optional: Thresholds for visualization
    thresholds: {
        good:     "< 100ms"
        warning:  "< 500ms"
        critical: ">= 500ms"
    }
}
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `description` | Yes | Human-readable explanation of what this signal measures |
| `metrics` | Yes | List of metric names (as defined in `metrics:`) that compose this signal |
| `recordingRules` | No | Pre-computed PromQL queries for dashboards and alerts |
| `thresholds` | No | Good/warning/critical thresholds for visualization |

## Recording Rules

Recording rules pre-compute expensive PromQL queries and store the result as a new metric. This provides:

1. **Performance** - Complex queries computed once, not on every dashboard refresh
2. **Consistency** - Same values across dashboards, alerts, and APIs
3. **Simplicity** - Use `http:server:latency:p99:5m` instead of the full query

### Naming Convention

Follow the Prometheus recording rule naming convention: `level:metric:operations`

```
http:server:latency:p99:5m
│    │      │       │   └── window (5 minutes)
│    │      │       └────── percentile
│    │      └────────────── signal type
│    └───────────────────── subsystem
└────────────────────────── namespace
```

### Examples by Signal Type

#### Latency Recording Rules

```cue
recordingRules: [
    {
        name:  "http:server:latency:p50:5m"
        query: "histogram_quantile(0.50, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
    },
    {
        name:  "http:server:latency:p95:5m"
        query: "histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
    },
    {
        name:  "http:server:latency:p99:5m"
        query: "histogram_quantile(0.99, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
    },
    {
        name:  "http:server:latency:p95:5m:by_path"
        query: "histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le, path))"
    },
]
```

#### Errors Recording Rules

```cue
recordingRules: [
    {
        name:  "http:server:errors:ratio:5m"
        query: "sum(rate(http_server_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_server_requests_total[5m]))"
    },
    {
        name:  "http:server:errors:rate:5m"
        query: "sum(rate(http_server_requests_total{status=~\"5..\"}[5m]))"
    },
]
```

#### Traffic Recording Rules

```cue
recordingRules: [
    {
        name:  "http:server:traffic:rate:5m"
        query: "sum(rate(http_server_requests_total[5m]))"
    },
    {
        name:  "http:server:traffic:rate:5m:by_method"
        query: "sum(rate(http_server_requests_total[5m])) by (method)"
    },
]
```

#### Saturation Recording Rules

```cue
recordingRules: [
    {
        name:  "http:server:saturation:in_flight"
        query: "sum(http_server_requests_in_flight)"
    },
    {
        name:  "db:postgres:saturation:pool_utilization"
        query: "db_postgres_connections_active / db_postgres_connections_max"
    },
]
```

## Thresholds

Thresholds define what values are considered good, warning, or critical. They appear in the HTML documentation and can be used to configure dashboard panels.

```cue
thresholds: {
    good:     "< 100ms"      // Green zone
    warning:  "< 500ms"      // Yellow zone (optional)
    critical: ">= 500ms"     // Red zone
}
```

The `warning` field is optional - some signals only have good/critical states.

### Threshold Examples

| Signal | Good | Warning | Critical |
|--------|------|---------|----------|
| Latency (P99) | < 100ms | < 500ms | >= 500ms |
| Error Rate | < 0.1% | < 1% | >= 1% |
| Saturation | < 70% | < 90% | >= 90% |

## Complete Example

Here's a comprehensive example with all four Golden Signals:

```cue
package main

version: "1.0.0"
info: {
    title:   "API Service"
    version: "1.0.0"
}
services: {
    api: {
        info: {
            title:   "API Service"
            version: "1.0.0"
        }
        metrics: {
            http_requests_total: {
                namespace: "http"
                subsystem: "server"
                type:      "counter"
                help:      "Total HTTP requests"
                labels: {
                    method: {description: "HTTP method"}
                    status: {description: "HTTP status code"}
                    path:   {description: "Request path"}
                }
            }
            http_request_duration_seconds: {
                namespace: "http"
                subsystem: "server"
                type:      "histogram"
                help:      "HTTP request duration in seconds"
                buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
                labels: {
                    method: {description: "HTTP method"}
                    path:   {description: "Request path"}
                }
            }
            http_requests_in_flight: {
                namespace: "http"
                subsystem: "server"
                type:      "gauge"
                help:      "Current number of HTTP requests being processed"
            }
        }

        goldenSignals: {
            "http/server": {
                latency: {
                    description: "How long HTTP requests take to complete"
                    metrics: ["http_request_duration_seconds"]
                    recordingRules: [
                        {
                            name:  "http:server:latency:p50:5m"
                            query: "histogram_quantile(0.50, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
                        },
                        {
                            name:  "http:server:latency:p95:5m"
                            query: "histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
                        },
                        {
                            name:  "http:server:latency:p99:5m"
                            query: "histogram_quantile(0.99, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
                        },
                    ]
                    thresholds: {
                        good:     "< 100ms"
                        warning:  "< 500ms"
                        critical: ">= 500ms"
                    }
                }
                errors: {
                    description: "Rate of failed HTTP requests (5xx responses)"
                    metrics: ["http_requests_total"]
                    recordingRules: [
                        {
                            name:  "http:server:errors:ratio:5m"
                            query: "sum(rate(http_server_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_server_requests_total[5m]))"
                        },
                    ]
                    thresholds: {
                        good:     "< 0.1%"
                        warning:  "< 1%"
                        critical: ">= 1%"
                    }
                }
                traffic: {
                    description: "Request volume - how much demand is placed on the service"
                    metrics: ["http_requests_total"]
                    recordingRules: [
                        {
                            name:  "http:server:traffic:rate:5m"
                            query: "sum(rate(http_server_requests_total[5m]))"
                        },
                        {
                            name:  "http:server:traffic:rate:5m:by_method"
                            query: "sum(rate(http_server_requests_total[5m])) by (method)"
                        },
                    ]
                }
                saturation: {
                    description: "How close to capacity the HTTP server is"
                    metrics: ["http_requests_in_flight"]
                    recordingRules: [
                        {
                            name:  "http:server:saturation:in_flight"
                            query: "sum(http_server_requests_in_flight)"
                        },
                    ]
                    thresholds: {
                        good:     "< 100"
                        warning:  "< 500"
                        critical: ">= 500"
                    }
                }
            }
        }
    }
}
```

## HTML Documentation

Golden Signals appear in the generated HTML documentation as a compact bar at the top of the page. Each signal shows:

- Signal name and icon
- Current "good" threshold value
- Expandable popover with:
  - Full description
  - All thresholds (good/warning/critical)
  - Recording rules with copy-to-clipboard

Generate HTML documentation with:

```bash
promener html -i metrics.cue -o docs/metrics.html
```

## Future: Recording Rules Export

A future version of Promener will support exporting recording rules to Prometheus-compatible YAML:

```bash
# Future command (not yet implemented)
promener rules -i metrics.cue -o prometheus/rules.yml
```

This will generate:

```yaml
groups:
  - name: http_server_golden_signals
    rules:
      - record: http:server:latency:p50:5m
        expr: histogram_quantile(0.50, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))
      - record: http:server:latency:p95:5m
        expr: histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))
      # ... more rules
```

## Best Practices

1. **Define all four signals** - Even if some don't have thresholds, document what metrics compose each signal

2. **Use consistent naming** - Follow the `level:metric:operations` convention for recording rules

3. **Include multiple percentiles** - For latency, define P50 (median), P95, and P99

4. **Match topics to metrics** - Use `namespace/subsystem` format that matches your metrics

5. **Set realistic thresholds** - Base thresholds on actual SLOs and historical data

6. **Document saturation carefully** - Choose the right saturation metric for your system (connections, memory, CPU, queue depth)

## Roadmap

The following features are planned for Golden Signals support:

- [x] **HTML Documentation** - Display Golden Signals in generated HTML documentation with interactive popovers
- [ ] **Recording Rules YAML Export** - Generate Prometheus-compatible recording rules YAML from Golden Signals definitions
- [ ] **Alerting Rules YAML Export** - Generate Prometheus alerting rules based on thresholds
- [ ] **Grafana Dashboard Generation** - Generate Grafana dashboard JSON with Golden Signals panels
- [ ] **AlertManager Config Generation** - Generate AlertManager routing and receiver configurations
- [ ] **Metric Reference Validation** - Validate that metrics referenced in Golden Signals exist in the service
- [ ] **Recording Rule Name Validation** - Validate recording rule names follow Prometheus naming conventions
- [ ] **Query Syntax Validation** - Validate PromQL queries in recording rules at specification time
- [ ] **Golden Signals Templates** - Pre-built Golden Signals definitions for common patterns (HTTP, gRPC, database, cache, queue)
- [ ] **SLO Integration** - Define SLOs based on Golden Signals with error budget calculations

## See Also

- [CUE Specification](cue-specification.md) - Full CUE specification reference
- [HTML Documentation](github-pages.md) - Publishing HTML docs
- [Google SRE Book - Monitoring](https://sre.google/sre-book/monitoring-distributed-systems/) - Original Golden Signals definition
