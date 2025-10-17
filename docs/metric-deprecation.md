# Metric Deprecation

Promener supports marking metrics as deprecated to help teams manage metric lifecycle and guide users toward replacement metrics. Deprecated metrics are clearly flagged in the generated HTML documentation with migration information.

## Table of Contents

- [When to Deprecate Metrics](#when-to-deprecate-metrics)
- [YAML Syntax](#yaml-syntax)
- [Deprecation Fields](#deprecation-fields)
- [Visual Indicators](#visual-indicators)
- [Best Practices](#best-practices)
- [Migration Strategy](#migration-strategy)
- [Examples](#examples)

## When to Deprecate Metrics

Deprecate metrics when:

- **Replacing with better metric types**: Moving from counter to histogram for latency tracking
- **Restructuring metric organization**: Changing namespace/subsystem structure
- **Renaming for clarity**: Improving metric naming conventions
- **Consolidating metrics**: Merging similar metrics into a unified one
- **Removing redundant metrics**: Eliminating metrics that duplicate data

**Important**: Don't delete deprecated metrics immediately. Keep them in the spec with deprecation information to help users migrate.

## YAML Syntax

Add the `deprecated` field to any metric definition:

```yaml
metrics:
  requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total number of HTTP requests"
    deprecated:
      since: "2024-01-15"
      replacedBy: "request_duration_seconds"
      reason: "Switching to histogram for better latency tracking"
    labels:
      - method
      - status
```

## Deprecation Fields

### `since` (optional)

When the metric was deprecated (date or version):

```yaml
deprecated:
  since: "2024-01-15"  # Date format: YYYY-MM-DD
```

or

```yaml
deprecated:
  since: "v2.0.0"  # Version format
```

### `replacedBy` (optional)

The name of the metric that replaces this one:

```yaml
deprecated:
  replacedBy: "request_duration_seconds"
```

Use the metric's short name, not the full name. For example, if the full metric name is `http_server_request_duration_seconds`, use `request_duration_seconds`.

### `reason` (optional)

A brief explanation of why the metric is deprecated:

```yaml
deprecated:
  reason: "Switching to histogram for better latency tracking"
```

## Visual Indicators

### In HTML Documentation

- **Sidebar**: Deprecated metrics show a ⚠️ warning icon next to the type badge
- **Metric details**: A prominent orange banner displays the deprecation information

### In Generated Code

Promener adds language-specific deprecation annotations to all generated methods:
- **Go**: `// Deprecated:` comments
- **.NET**: `[Obsolete]` attributes
- **Node.js**: `@deprecated` JSDoc tags

Your IDE will show warnings (strikethrough text, hover messages) when you use deprecated metrics.

## Best Practices

### 1. Always Provide Migration Information

Include all three fields when deprecating a metric:

```yaml
# ✅ GOOD: Complete deprecation info
deprecated:
  since: "2024-01-15"
  replacedBy: "request_duration_seconds"
  reason: "Switching to histogram for better latency tracking"

# ❌ BAD: Missing critical information
deprecated:
  since: "2024-01-15"
```

### 2. Keep Deprecated Metrics for a Grace Period

Don't remove deprecated metrics immediately. Keep them for at least:
- **3-6 months** for internal services
- **12+ months** for public APIs
- **Until next major version** for libraries

```yaml
# Keep both metrics during transition period
metrics:
  # Deprecated metric
  requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total number of HTTP requests"
    deprecated:
      since: "2024-01-15"
      replacedBy: "request_duration_seconds"
      reason: "Switching to histogram for better latency tracking"
    labels:
      - method
      - status

  # Replacement metric
  request_duration_seconds:
    namespace: http
    subsystem: server
    type: histogram
    help: "HTTP request duration in seconds"
    labels:
      - method
      - endpoint
    buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
```

### 3. Use Clear, Actionable Reasons

Explain **why** the metric is deprecated and **what to do instead**:

```yaml
# ✅ GOOD: Clear and actionable
reason: "Switching to histogram for better latency tracking. Use request_duration_seconds with bucket queries."

# ❌ BAD: Vague
reason: "Old metric"
```

### 4. Document the Migration Path

Include examples showing how to migrate queries:

```yaml
deprecated:
  since: "2024-01-15"
  replacedBy: "request_duration_seconds"
  reason: "Switching to histogram for better latency tracking"
examples:
  promql:
    - query: 'rate(http_server_requests_total[5m])'
      description: "OLD: Request rate (deprecated)"
    - query: 'rate(http_server_request_duration_seconds_count[5m])'
      description: "NEW: Request rate using histogram _count"
```

### 5. Use Consistent Date Format

Use ISO 8601 date format (YYYY-MM-DD) for consistency:

```yaml
# ✅ GOOD
since: "2024-01-15"

# ❌ INCONSISTENT
since: "Jan 15, 2024"
since: "15/01/2024"
```

## Migration Strategy

### Phase 1: Add New Metric

Introduce the replacement metric alongside the old one:

```yaml
metrics:
  # Existing metric (will be deprecated)
  api_calls:
    namespace: api
    subsystem: gateway
    type: counter
    help: "Total API calls"

  # New metric
  api_call_duration_seconds:
    namespace: api
    subsystem: gateway
    type: histogram
    help: "API call duration in seconds"
    buckets: [0.01, 0.05, 0.1, 0.5, 1, 5]
```

### Phase 2: Mark as Deprecated

After the new metric is deployed and working:

```yaml
metrics:
  api_calls:
    namespace: api
    subsystem: gateway
    type: counter
    help: "Total API calls"
    deprecated:
      since: "2024-03-01"
      replacedBy: "api_call_duration_seconds"
      reason: "Histogram provides better latency insights"
```

### Phase 3: Monitor Adoption

Track usage of both metrics:

```promql
# Check if old metric is still being queried
prometheus_api_v1_query_range_total{query=~".*api_gateway_api_calls.*"}
```

### Phase 4: Remove After Grace Period

After the grace period and when usage drops to zero:

1. Update dashboards to use new metric
2. Update alerts to use new metric
3. Remove the deprecated metric from the YAML spec

## Examples

### Example 1: Counter to Histogram

```yaml
metrics:
  # Deprecated counter
  http_requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total HTTP requests"
    deprecated:
      since: "2024-01-15"
      replacedBy: "http_request_duration_seconds"
      reason: "Histogram provides request counts plus latency percentiles"
    labels:
      - method
      - status
    examples:
      promql:
        - query: 'rate(http_server_http_requests_total[5m])'
          description: "OLD: Request rate"
        - query: 'rate(http_server_http_request_duration_seconds_count[5m])'
          description: "NEW: Request rate from histogram _count"

  # New histogram
  http_request_duration_seconds:
    namespace: http
    subsystem: server
    type: histogram
    help: "HTTP request duration in seconds"
    labels:
      - method
      - endpoint
    buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
```

### Example 2: Renaming for Clarity

```yaml
metrics:
  # Old unclear name
  queue_size:
    namespace: worker
    subsystem: jobs
    type: gauge
    help: "Number of pending jobs"
    deprecated:
      since: "2024-02-01"
      replacedBy: "pending_jobs"
      reason: "Renaming for clarity - 'pending_jobs' is more descriptive"

  # New clear name
  pending_jobs:
    namespace: worker
    subsystem: jobs
    type: gauge
    help: "Number of pending jobs in the worker queue"
    labels:
      - priority
      - queue_name
```

### Example 3: Namespace Restructuring

```yaml
metrics:
  # Old structure
  db_queries:
    namespace: app
    subsystem: database
    type: counter
    help: "Total database queries"
    deprecated:
      since: "2024-03-01"
      replacedBy: "queries_total"
      reason: "Reorganizing metrics under 'db' namespace for consistency"
    labels:
      - operation

  # New structure
  queries_total:
    namespace: db
    subsystem: postgres
    type: counter
    help: "Total PostgreSQL queries"
    labels:
      - operation
      - table
```

### Example 4: Consolidating Metrics

```yaml
metrics:
  # Deprecated separate metrics
  http_get_requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total HTTP GET requests"
    deprecated:
      since: "2024-01-10"
      replacedBy: "requests_total"
      reason: "Consolidating method-specific metrics into requests_total with method label"

  http_post_requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total HTTP POST requests"
    deprecated:
      since: "2024-01-10"
      replacedBy: "requests_total"
      reason: "Consolidating method-specific metrics into requests_total with method label"

  # New consolidated metric
  requests_total:
    namespace: http
    subsystem: server
    type: counter
    help: "Total HTTP requests"
    labels:
      - method
      - status
      - endpoint
```

## Related Documentation

- [YAML Specification](yaml-specification.md)
- [Generated Code Structure](generated-code.md)
- [Constant Labels](constant-labels.md)
- [HTTP Server Integration](http-integration.md)
