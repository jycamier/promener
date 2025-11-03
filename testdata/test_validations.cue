package main

version: "1.0.0"
info: {
	title:   "Test Metrics with Validations"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title:   "Default Service"
			version: "1.0.0"
		}
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
							"value in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']",
						]
					}
					status: {
						description: "HTTP status code"
						validations: [
							"value.matches('^[1-5][0-9]{2}$')",
						]
					}
					service: {
						description: "Service name"
						validations: [
							"value.matches('^[a-z][a-z0-9-]*$')",
							"size(value) >= 3",
							"size(value) <= 63",
						]
					}
				}
			}
		}
	}
}
