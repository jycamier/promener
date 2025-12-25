package PromenerPolicy

# Histogram and Summary naming
# They should generally end in the unit (e.g., _seconds)
PromenerPolicy contains result if {
    metric := get_metrics_common[_]
    metric.type == "histogram"
    
    valid_suffixes := ["_seconds", "_bytes"]
    not ends_with_any(metric.full_name, valid_suffixes)

    result := {
        "path": metric.path,
        "message": sprintf("Histogram '%s' should end with a unit suffix (e.g., _seconds, _bytes)", [metric.full_name]),
        "severity": "error"
    }
}

ends_with_any(str, suffixes) if {
    endswith(str, suffixes[_])
}
