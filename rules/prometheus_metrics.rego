package PromenerPolicy

# Counters MUST have _total suffix
PromenerPolicy contains result if {
    metric := get_metrics_common[_]
    metric.type == "counter"
    not endswith(metric.full_name, "_total")

    result := {
        "severity": "error",
        "message": sprintf("Counter metric '%s' should end with '_total'", [metric.full_name]),
        "path": metric.path
    }
}

# Non-counters SHOULD NOT have _total suffix
PromenerPolicy contains result if {
    metric := get_metrics_common[_]
    metric.type != "counter"
    endswith(metric.full_name, "_total")

    result := {
        "severity": "warning",
        "message": sprintf("Non-counter metric '%s' (type: %s) should not end with '_total'", [metric.full_name, metric.type]),
        "path": metric.path
    }
}

# Base units should be plural
PromenerPolicy contains result if {
    forbidden_units := ["_second", "_byte", "_gram", "_meter", "_bit"]
    metric := get_metrics_common[_]
    unit := forbidden_units[_]
    endswith(metric.full_name, unit)

    result := {
        "severity": "error",
        "message": sprintf("Metric '%s' should use plural unit (e.g. %ss)", [metric.full_name, unit]),
        "path": metric.path
    }
}

# Avoid repeated words in metric names (e.g. http_server_server_error)
PromenerPolicy contains result if {
    metric := get_metrics_common[_]
    parts := split(metric.full_name, "_")
    
    some i
    # Check if the next part exists and matches the current part
    parts[i] == parts[i+1]

    result := {
        "severity": "warning",
        "message": sprintf("Metric name '%s' contains repeated segment '%s'", [metric.full_name, parts[i]]),
        "path": metric.path
    }
}