package main

import (
	"github.com/jycamier/promener-standards/primitives/labels"
	"github.com/jycamier/promener-standards/metrics/latency"
	"github.com/jycamier/promener-standards/metrics/traffic"
)

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
			#commonLabels: labels.#C4Labels
		}
		traffic.#DefaultTrafficMetrics & {
			#commonLabels: labels.#C4Labels
		}
		metrics: {
			test_test: {
				namespace: "test"
				subsystem: "test"
				help:      "mon help"
				type:      "gauge"
			}
		}
	}
}
