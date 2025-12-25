package PromenerPolicy

# Helper to construct the full metric name
get_full_name(metric, key) := name if {
    metric.name
    part := metric.name
    name := sprintf("%s_%s_%s", [metric.namespace, metric.subsystem, part])
}

get_full_name(metric, key) := name if {
    not metric.name
    name := sprintf("%s_%s_%s", [metric.namespace, metric.subsystem, key])
}

# Common helper to iterate over metrics with enriched data
get_metrics_common contains res if {
    some service_name, key
    service := input.services[service_name]
    metric := service.metrics[key]
    
    full_name := get_full_name(metric, key)
    
    # Handle optional labels safely
    labels := object.get(metric, "labels", {})

    res := {
        "service_name": service_name,
        "key": key,
        "full_name": full_name,
        "type": metric.type,
        "labels": labels,
        "path": sprintf("services[%s].metrics[%s]", [service_name, key])
    }
}