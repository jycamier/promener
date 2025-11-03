package main

import (
	"github.com/jycamier/promener-standards/primitives/labels"
	"github.com/jycamier/promener-standards/primitives/histogram"
	"github.com/jycamier/promener-standards/metrics/latency"
	"github.com/jycamier/promener-standards/metrics/traffic"
)

let commonLabels = labels.#C4Labels

version: "1.0.0"
info: {
	title:   "Ecommerce 9000"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title:   "Default Service"
			version: "1.0.0"
		}
		latency.#DefaultLatencyMetrics & {
			#commonLabels: commonLabels
		}
		traffic.#DefaultTrafficMetrics & {
			#commonLabels: commonLabels
		}
		metrics: {
			old_metric: {
				namespace: "http"
				subsystem: "request"
				type:      "histogram"
				help:      "Old metric - use new_metric instead"
				deprecated: {
					since:      "2024-01-15"
					replacedBy: "foo_foobar_new_metric"
					reason:     "Switching to histogram for better latency tracking"
				}
				buckets: histogram.#HTTPBuckets
				labels: commonLabels & {
					method: {
						description: "HTTP method"
						validations: [
							"value in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']",
						]
					}
				}
			}
			new_metric: {
				namespace: "http"
				subsystem: "request"
				type:      "histogram"
				help:      "New metric - use new_metric instead"
				buckets: histogram.#HTTPBuckets
				labels: commonLabels & {
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
			}
		}
	}
}
