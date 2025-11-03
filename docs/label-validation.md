# Label Validation with CEL

Promener supports runtime label validation using [CEL (Common Expression Language)](https://github.com/google/cel-spec). This ensures that label values meet your requirements before being recorded to Prometheus, preventing invalid metrics and catching bugs early.

## Overview

Label validation provides:
- **Runtime safety**: Validates label values before recording metrics
- **Expressive constraints**: Use CEL for powerful validation rules
- **Fast performance**: Validations compiled once at initialization
- **Clear error messages**: Descriptive panics when validation fails
- **Optional validation**: Add validations only where needed

## Why CEL?

CEL (Common Expression Language) is:
- **Safe**: No side effects, deterministic, and terminates
- **Fast**: Compiles expressions for efficient runtime evaluation
- **Expressive**: Rich set of operators and functions
- **Standard**: Used by Kubernetes, Envoy, and other projects
- **Type-safe**: Strong typing with automatic type checking

## Basic Usage

Add validations to your label definitions in CUE:

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
}
```

In the generated code, validations run before each metric observation:

```go
// Valid - passes validation
metrics.Http.Server.IncRequestsTotal("GET", "200")

// Panics: label "method" value "INVALID" failed validation:
//   value in ["GET", "POST", "PUT", "DELETE", "PATCH"] (false)
metrics.Http.Server.IncRequestsTotal("INVALID", "200")
```

## Validation Expressions

### Enum Validation

Validate that a value is one of a set of allowed values:

```cue
method: {
    description: "HTTP method"
    validations: [
        "value in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']",
    ]
}
```

### Regex Matching

Use regular expressions to validate string patterns:

```cue
status: {
    description: "HTTP status code"
    validations: [
        "value.matches('^[1-5][0-9]{2}$')",
    ]
}

uuid: {
    description: "Request UUID"
    validations: [
        "value.matches('^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')",
    ]
}
```

### String Prefix/Suffix

Check if strings start or end with specific patterns:

```cue
service: {
    description: "Service name"
    validations: [
        "value.startsWith('api-')",
    ]
}

endpoint: {
    description: "API endpoint"
    validations: [
        "value.startsWith('/api/')",
    ]
}

version: {
    description: "API version"
    validations: [
        "value.startsWith('v')",
        "value.matches('^v[0-9]+$')",
    ]
}
```

### Length Constraints

Validate string length using `size()`:

```cue
service: {
    description: "Service name (3-63 characters)"
    validations: [
        "size(value) >= 3",
        "size(value) <= 63",
    ]
}

correlation_id: {
    description: "Correlation ID (exactly 32 characters)"
    validations: [
        "size(value) == 32",
    ]
}
```

### Combining Constraints

Use multiple validations for complex requirements:

```cue
service: {
    description: "Service name (DNS-compatible)"
    validations: [
        "value.matches('^[a-z][a-z0-9-]*$')",  // Lowercase, starts with letter
        "size(value) >= 3",                     // Minimum length
        "size(value) <= 63",                    // Maximum length
        "!value.endsWith('-')",                 // No trailing hyphen
    ]
}

pod_name: {
    description: "Kubernetes pod name"
    validations: [
        "value.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",
        "size(value) <= 253",
    ]
}
```

### Logical Operators

Use `&&`, `||`, and `!` for complex logic:

```cue
environment: {
    description: "Deployment environment"
    validations: [
        "value in ['dev', 'staging', 'prod'] || value.startsWith('test-')",
    ]
}

label: {
    description: "Label value"
    validations: [
        "size(value) > 0 && size(value) <= 100",
        "!value.contains('  ')",  // No double spaces
    ]
}
```

## CEL Built-in Functions

### String Functions

| Function | Description | Example |
|----------|-------------|---------|
| `size(value)` | String length | `size(value) >= 3` |
| `value.matches(pattern)` | Regex match | `value.matches('^[a-z]+$')` |
| `value.startsWith(prefix)` | Starts with | `value.startsWith('api-')` |
| `value.endsWith(suffix)` | Ends with | `value.endsWith('-svc')` |
| `value.contains(substr)` | Contains substring | `value.contains('test')` |

### Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equality | `value == 'prod'` |
| `!=` | Inequality | `value != ''` |
| `in` | Membership | `value in ['a', 'b', 'c']` |
| `&&` | Logical AND | `size(value) >= 3 && size(value) <= 63` |
| `\|\|` | Logical OR | `value == 'dev' \|\| value == 'prod'` |
| `!` | Logical NOT | `!value.startsWith('_')` |
| `<`, `<=`, `>`, `>=` | Comparison | `size(value) >= 1` |

## Complete Examples

### HTTP Metrics

```cue
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
                    "value in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']",
                ]
            }
            status: {
                description: "HTTP status code"
                validations: [
                    "value.matches('^[1-5][0-9]{2}$')",
                ]
            }
            path: {
                description: "Request path"
                validations: [
                    "value.startsWith('/')",
                    "size(value) >= 1",
                    "size(value) <= 256",
                ]
            }
        }
    }
}
```

### Database Metrics

```cue
metrics: {
    db_queries_total: {
        namespace: "db"
        subsystem: "postgres"
        type:      "counter"
        help:      "Total database queries"
        labels: {
            operation: {
                description: "SQL operation type"
                validations: [
                    "value in ['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'BEGIN', 'COMMIT', 'ROLLBACK']",
                ]
            }
            table: {
                description: "Table name"
                validations: [
                    "value.matches('^[a-z_][a-z0-9_]*$')",  // Valid SQL identifier
                    "size(value) >= 1",
                    "size(value) <= 63",
                ]
            }
            result: {
                description: "Query result (success or error)"
                validations: [
                    "value in ['success', 'error']",
                ]
            }
        }
    }
}
```

### Kubernetes Metrics

```cue
metrics: {
    k8s_pod_restarts_total: {
        namespace: "k8s"
        subsystem: "pod"
        type:      "counter"
        help:      "Total pod restarts"
        labels: {
            namespace: {
                description: "Kubernetes namespace"
                validations: [
                    "value.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",
                    "size(value) >= 1",
                    "size(value) <= 63",
                ]
            }
            pod: {
                description: "Pod name"
                validations: [
                    "value.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",
                    "size(value) <= 253",
                ]
            }
            reason: {
                description: "Restart reason"
                validations: [
                    "value in ['Error', 'Completed', 'OOMKilled', 'CrashLoopBackOff']",
                ]
            }
        }
    }
}
```

## Generated Code

For a metric with validations:

```cue
labels: {
    method: {
        description: "HTTP method"
        validations: [
            "value in ['GET', 'POST', 'PUT', 'DELETE']",
        ]
    }
}
```

Promener generates Go code that:

1. **Compiles validations at initialization** (one-time cost):

```go
func (m *httpServerMetrics) init() {
    m.once.Do(func() {
        // Compile CEL programs
        env, _ := cel.NewEnv(cel.Variable("value", cel.StringType))

        // Compile "method" validations
        methodProgram0, _ := env.Compile(`value in ["GET", "POST", "PUT", "DELETE"]`)
        m.requestsTotalMethodValidation0, _ = env.Program(methodProgram0)

        // Create metric
        m.requestsTotal = prometheus.NewCounterVec(...)
        prometheus.MustRegister(m.requestsTotal)
    })
}
```

2. **Validates before each observation**:

```go
func (m *httpServerMetrics) IncRequestsTotal(method string, status string) {
    m.init()

    // Validate "method" label
    if m.requestsTotalMethodValidation0 != nil {
        result, _ := m.requestsTotalMethodValidation0.Eval(map[string]interface{}{
            "value": method,
        })
        if result.Value() != true {
            panic(fmt.Sprintf(`label "method" value %q failed validation: %s`,
                method, `value in ["GET", "POST", "PUT", "DELETE"]`))
        }
    }

    // Record metric
    m.requestsTotal.WithLabelValues(method, status).Inc()
}
```

## Performance Considerations

### Compilation Cost

CEL expressions are compiled **once** during metric initialization:
- Compilation happens in `sync.Once`, ensuring thread-safety
- Programs are reused for all metric observations
- No parsing overhead at observation time

### Runtime Cost

Runtime validation is fast:
- Pre-compiled CEL programs execute in microseconds
- Simple expressions (enum checks) are nearly free
- Regex matching is more expensive but still sub-millisecond
- No allocations in the hot path for simple validations

### When to Use Validations

Add validations when:
- ✅ Label values come from untrusted sources (user input, external APIs)
- ✅ Invalid labels would cause monitoring confusion
- ✅ You want to catch bugs during development
- ✅ Label values must follow strict formats (status codes, UUIDs)

Skip validations when:
- ❌ Label values are compile-time constants
- ❌ Performance is extremely critical (millions of observations/second)
- ❌ Labels are already validated elsewhere in your code

## Error Messages

When validation fails, Promener panics with a descriptive message:

```
panic: label "method" value "INVALID" failed validation: value in ["GET", "POST", "PUT", "DELETE"] (false)
```

The message includes:
1. **Label name**: Which label failed (`method`)
2. **Invalid value**: What value was provided (`INVALID`)
3. **Validation expression**: The CEL expression that failed
4. **Result**: The boolean result of the expression

This makes debugging easy during development and testing.

## Validation at Build Time

Use the `vet` command to validate your CUE specifications, including CEL expressions:

```bash
promener vet metrics.cue
```

This checks:
- ✅ CEL syntax is correct
- ✅ CEL expressions compile successfully
- ✅ Validations reference the `value` variable
- ✅ Expressions return boolean results

See [Vet Command](vet-command.md) for more details.

## Best Practices

1. **Validate untrusted input**: Always validate labels from external sources
2. **Use enums for known values**: Prefer `in` operator for fixed sets
3. **Keep validations simple**: Complex logic can be split across multiple validations
4. **Test your validations**: Write unit tests that try invalid label values
5. **Document constraints**: Use label descriptions to explain validation rules
6. **Fail fast**: Let validations panic during development to catch bugs early
7. **Be specific with regex**: Use `^` and `$` anchors for full string matching

## Testing with Validations

When testing code with validated metrics, expect panics for invalid labels:

```go
func TestMetricValidation(t *testing.T) {
    metrics := metrics.Default()

    // Valid - should not panic
    metrics.Http.Server.IncRequestsTotal("GET", "200")

    // Invalid - should panic
    assert.Panics(t, func() {
        metrics.Http.Server.IncRequestsTotal("INVALID", "200")
    })
}
```

For testing without validations, use mock implementations or test-specific metrics without validation rules.

## Troubleshooting

### Validation Always Fails

**Problem**: Validation expression returns false for all values.

**Solution**: Check CEL syntax and operators:
- Use `==` for equality, not `=`
- Strings must be in quotes: `'GET'`
- Use `in` for lists: `value in ['a', 'b']`

### Panic During Initialization

**Problem**: CEL expression fails to compile.

**Solution**: Run `promener vet` to check syntax:
```bash
promener vet metrics.cue
```

Common issues:
- Missing quotes around strings
- Invalid regex patterns
- Undefined functions

### Performance Issues

**Problem**: High CPU usage from validations.

**Solution**:
- Profile your application to confirm validations are the bottleneck
- Simplify complex regex patterns
- Remove validations from hot paths
- Cache validated values if possible

## See Also

- [CUE Specification](cue-specification.md) - Complete CUE format reference
- [Vet Command](vet-command.md) - Validating specifications
- [CEL Language](https://github.com/google/cel-spec) - Official CEL documentation
- [CEL Go Library](https://github.com/google/cel-go) - Go implementation used by Promener
