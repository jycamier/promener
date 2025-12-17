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
			requests_in_flight: {
				namespace: "http"
				subsystem: "request"
				type:      "gauge"
				help:      "Current number of HTTP requests being processed"
				labels: commonLabels
			}
		}

		goldenSignals: {
			"http/request": {
				latency: {
					description: "HTTP request latency distribution"
					metrics: ["new_metric"]
					recordingRules: [
						{
							name:  "http:request:latency:p50:5m"
							query: "histogram_quantile(0.50, sum(rate(http_request_new_metric_bucket[5m])) by (le))"
						},
						{
							name:  "http:request:latency:p95:5m"
							query: "histogram_quantile(0.95, sum(rate(http_request_new_metric_bucket[5m])) by (le))"
						},
						{
							name:  "http:request:latency:p99:5m"
							query: "histogram_quantile(0.99, sum(rate(http_request_new_metric_bucket[5m])) by (le))"
						},
					]
					thresholds: {
						good:     "< 100ms"
						warning:  "< 500ms"
						critical: ">= 500ms"
					}
				}
				errors: {
					description: "HTTP error rate (5xx responses)"
					metrics: ["new_metric"]
					recordingRules: [
						{
							name:  "http:request:errors:ratio:5m"
							query: "sum(rate(http_request_new_metric_count{status=~\"5..\"}[5m])) / sum(rate(http_request_new_metric_count[5m]))"
						},
					]
					thresholds: {
						good:     "< 0.1%"
						warning:  "< 1%"
						critical: ">= 1%"
					}
				}
				traffic: {
					description: "HTTP request throughput"
					metrics: ["new_metric"]
					recordingRules: [
						{
							name:  "http:request:traffic:rate:5m"
							query: "sum(rate(http_request_new_metric_count[5m]))"
						},
						{
							name:  "http:request:traffic:rate:5m:by_method"
							query: "sum(rate(http_request_new_metric_count[5m])) by (method)"
						},
					]
				}
				saturation: {
					description: "Current load on the HTTP server - high values indicate potential bottlenecks"
					metrics: ["requests_in_flight"]
					recordingRules: [
						{
							name:  "http:request:saturation:in_flight"
							query: "sum(http_request_requests_in_flight)"
						},
						{
							name:  "http:request:saturation:in_flight:max_1h"
							query: "max_over_time(sum(http_request_requests_in_flight)[1h:])"
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
