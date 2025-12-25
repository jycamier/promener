package PromenerPolicy

# Avoid labels in metric names
PromenerPolicy contains result if {
    metric := get_metrics_common[_]
    metric.labels[label_name]
    contains(metric.full_name, label_name)

    result := {
        "path": sprintf("%s.labels[%s]", [metric.path, label_name]),
        "message": sprintf("Metric name '%s' should not contain label name '%s'", [metric.full_name, label_name]),
        "severity": "warning"
    }
}

# Reserved labels (job, instance)
PromenerPolicy contains result if {
    reserved_labels := {"job", "instance"}
    metric := get_metrics_common[_]
    metric.labels[label_name]
    reserved_labels[label_name]

    result := {
        "path": sprintf("%s.labels[%s]", [metric.path, label_name]),
        "message": sprintf("Label '%s' is reserved by Prometheus", [label_name]),
        "severity": "error"
    }
}