package PromenerPolicy

# Test valid histogram with _seconds suffix
test_histogram_seconds_valid if {
    mock_input := {
        "services": {
            "api": {
                "metrics": {
                    "request_duration": {
                        "namespace": "http",
                        "subsystem": "server",
                        "type": "histogram",
                        "name": "request_duration_seconds"
                    }
                }
            }
        }
    }

    count(PromenerPolicy) == 0 with input as mock_input
}

# Test valid histogram with _bytes suffix
test_histogram_bytes_valid if {
    mock_input := {
        "services": {
            "api": {
                "metrics": {
                    "response_size": {
                        "namespace": "http",
                        "subsystem": "server",
                        "type": "histogram",
                        "name": "response_size_bytes"
                    }
                }
            }
        }
    }

    count(PromenerPolicy) == 0 with input as mock_input
}

# Test invalid histogram missing unit suffix
test_histogram_no_unit_invalid if {
    mock_input := {
        "services": {
            "api": {
                "metrics": {
                    "request_latency": {
                        "namespace": "http",
                        "subsystem": "server",
                        "type": "histogram"
                        # full_name will be http_server_request_latency (no unit)
                    }
                }
            }
        }
    }

    results := PromenerPolicy with input as mock_input
    count(results) == 1
    result := results[_]
    result.severity == "error"
    contains(result.message, "should end with a unit suffix")
}

# Test invalid histogram using _total suffix (which was removed from valid list)
test_histogram_total_invalid if {
    mock_input := {
        "services": {
            "api": {
                "metrics": {
                    "request_total": {
                        "namespace": "http",
                        "subsystem": "server",
                        "type": "histogram"
                    }
                }
            }
        }
    }

    results := PromenerPolicy with input as mock_input
    # This should trigger BOTH:
    # 1. Histogram rule (error: no _seconds/_bytes)
    # 2. Metric rule (warning: non-counter with _total)
    count(results) >= 1
    
    # Check for the histogram specific error
    some i
    results[i].severity == "error"
    contains(results[i].message, "should end with a unit suffix")
}