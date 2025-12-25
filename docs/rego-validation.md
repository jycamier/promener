# Policy Validation with Rego

Promener supports advanced validation policies using [OPA (Open Policy Agent)](https://www.openpolicyagent.org/) Rego language. This allows you to enforce organizational standards, naming conventions, and best practices beyond the basic schema validation.

## Getting Started

1. Create a directory for your rules (e.g., `rules/`).
2. Write your policies in `.rego` files inside that directory.
3. Run Promener with the `--rules` flag or configure it in `.promener.yaml`.

## Writing Rules

Promener expects Rego policies to be in the `PromenerPolicy` package and to return a set of result objects.

### Input Structure

The `input` document available to your rules mirrors the JSON structure of your CUE specification:

```json
{
  "services": {
    "my-service": {
      "metrics": {
        "http_requests_total": {
          "type": "counter",
          "labels": { ... }
        }
      }
    }
  }
}
```

### Result Format

Your rules should return objects with the following fields:
- `message`: A human-readable description of the violation.
- `severity`: The severity level (`error`, `warning`, or `info`). Defaults to `error`.
- `path` (optional): The path to the invalid element, used for reporting.

### Example Rule

Here is a rule that enforces a naming convention for counters:

```rego
package PromenerPolicy

# Helper to iterate over metrics
get_metrics[res] {
    some service_name, metric_name
    service := input.services[service_name]
    metric := service.metrics[metric_name]
    res := {
        "service_name": service_name,
        "metric_name": metric_name,
        "type": metric.type,
        "path": sprintf("services[%s].metrics[%s]", [service_name, metric_name])
    }
}

# Rule: Counters must end with _total
PromenerPolicy[result] {
    metric := get_metrics[_]
    metric.type == "counter"
    not endswith(metric.metric_name, "_total")

    result := {
        "severity": "error",
        "message": sprintf("Counter '%s' must end with '_total'", [metric.metric_name]),
        "path": metric.path
    }
}
```

## Running Validation

Use the `vet` command to check your specifications:

```bash
promener vet metrics.cue --rules ./rules
```

Or configure it in `.promener.yaml`:

```yaml
rules: ./rules
```

## Severity Levels

You can control which severity level causes the validation to fail (exit code 1) using the `--severity-on-error` flag.

- `error` (default): Only errors cause failure. Warnings are displayed but do not break the build.
- `warning`: Errors and warnings cause failure.
- `info`: Any violation causes failure.

```bash
# Fail on warnings or errors
promener vet metrics.cue --rules ./rules --severity-on-error warning
```

## Testing Your Rules

Test files must use the `*_test.rego` naming convention and test rules must start with `test_`:

```rego
package PromenerPolicy

test_histogram_valid if {
    mock_input := {"services": {"api": {"metrics": {"duration": {
        "namespace": "http", "subsystem": "server",
        "type": "histogram", "name": "duration_seconds"
    }}}}}
    count(PromenerPolicy) == 0 with input as mock_input
}

test_histogram_invalid if {
    mock_input := {"services": {"api": {"metrics": {"latency": {
        "namespace": "http", "subsystem": "server", "type": "histogram"
    }}}}}
    count(PromenerPolicy) == 1 with input as mock_input
}
```

Run tests with:

```bash
opa test ./rules/ -v
```

## Built-in Best Practices

We recommend implementing the following [Prometheus naming best practices](https://prometheus.io/docs/practices/naming/):

- **Counters**: Must end with `_total`.
- **Units**: Should be plural (e.g., `_seconds`, `_bytes`) and base units.
- **Histograms**: Should end with a unit suffix.
- **Labels**: Should not be included in the metric name.
- **Reserved Labels**: Do not use `job` or `instance`.
