package main

version: "1.0.0"

info: {
	title:       "Order Service Observability"
	description: "Complete observability specification for the Order Service, including metrics, golden signals, and operational documentation."
	version:     "1.0.0"
}

services: {
	order_service: {
		info: {
			title:       "Order Service"
			description: "Handles order creation, payment processing, and fulfillment tracking."
			version:     "1.0.0"
		}

		servers: [
			{url: "http://localhost:9090", description: "Local Prometheus"},
			{url: "https://prometheus.prod.internal", description: "Production Prometheus"},
			{url: "https://grafana.prod.internal/d/order-service", description: "Grafana Dashboard"},
		]

		metrics: {
			// =========================================
			// HTTP metrics
			// =========================================
			http_requests_total: {
				namespace: "http"
				subsystem: "server"
				type:      "counter"
				help:      "Total number of HTTP requests processed by the order service"
				labels: {
					method: {
						description: "HTTP method (GET, POST, PUT, DELETE, PATCH)"
						validations: [
							"value in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']",
						]
					}
					status: {
						description: "HTTP response status code"
						validations: [
							"value.matches('^[1-5][0-9]{2}$')",
						]
					}
					path: {
						description: "Request path pattern (e.g., /api/v1/orders/{id})"
						validations: [
							"value.startsWith('/')",
							"size(value) <= 200",
						]
					}
					service: {
						description: "Calling service name"
						inherited:   "Automatically injected by the service mesh (Istio/Linkerd)"
					}
				}
				constLabels: {
					app: {
						value:       "order-service"
						description: "Application identifier"
					}
					env: {
						value:       "${ENVIRONMENT:production}"
						description: "Deployment environment"
					}
				}
				examples: {
					promql: [
						{
							query:       "sum(rate(http_server_requests_total[5m])) by (status)"
							description: "Request rate by status code"
						},
						{
							query:       "sum(rate(http_server_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_server_requests_total[5m]))"
							description: "Error rate (5xx responses)"
						},
						{
							query:       "topk(10, sum(rate(http_server_requests_total[5m])) by (path))"
							description: "Top 10 endpoints by request rate"
						},
					]
					alerts: [
						{
							name:        "HighErrorRate"
							expr:        "sum(rate(http_server_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_server_requests_total[5m])) > 0.01"
							description: "Error rate exceeds 1% of total requests"
							for:         "5m"
							severity:    "critical"
						},
						{
							name:        "HighLatencyP99"
							expr:        "histogram_quantile(0.99, rate(http_server_request_duration_seconds_bucket[5m])) > 1"
							description: "P99 latency exceeds 1 second"
							for:         "10m"
							severity:    "warning"
						},
					]
				}
			}

			http_request_duration_seconds: {
				namespace: "http"
				subsystem: "server"
				type:      "histogram"
				help:      "HTTP request duration in seconds"
				buckets: [0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
				labels: {
					method: {
						description: "HTTP method"
						validations: ["value in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']"]
					}
					path: {
						description: "Request path pattern"
					}
					status: {
						description: "HTTP response status code"
					}
				}
				examples: {
					promql: [
						{
							query:       "histogram_quantile(0.50, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
							description: "Median (P50) latency"
						},
						{
							query:       "histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
							description: "P95 latency"
						},
						{
							query:       "histogram_quantile(0.99, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
							description: "P99 latency"
						},
						{
							query:       "histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le, path))"
							description: "P95 latency by endpoint"
						},
					]
				}
			}

			http_requests_in_flight: {
				namespace: "http"
				subsystem: "server"
				type:      "gauge"
				help:      "Current number of HTTP requests being processed"
				examples: {
					promql: [
						{
							query:       "http_server_requests_in_flight"
							description: "Current concurrent requests"
						},
						{
							query:       "max_over_time(http_server_requests_in_flight[1h])"
							description: "Peak concurrent requests in the last hour"
						},
					]
					alerts: [
						{
							name:        "HighConcurrency"
							expr:        "http_server_requests_in_flight > 500"
							description: "Too many concurrent requests, may indicate a bottleneck"
							for:         "5m"
							severity:    "warning"
						},
					]
				}
			}

			http_request_size_bytes: {
				namespace: "http"
				subsystem: "server"
				type:      "histogram"
				help:      "HTTP request body size in bytes"
				buckets: [100, 1000, 10000, 100000, 1000000, 10000000]
				labels: {
					method: {description: "HTTP method"}
					path:   {description: "Request path pattern"}
				}
			}

			http_response_size_bytes: {
				namespace: "http"
				subsystem: "server"
				type:      "histogram"
				help:      "HTTP response body size in bytes"
				buckets: [100, 1000, 10000, 100000, 1000000, 10000000]
				labels: {
					method: {description: "HTTP method"}
					path:   {description: "Request path pattern"}
					status: {description: "HTTP response status code"}
				}
			}

			// =========================================
			// Database metrics
			// =========================================
			db_queries_total: {
				namespace: "db"
				subsystem: "postgres"
				type:      "counter"
				help:      "Total number of database queries executed"
				labels: {
					operation: {
						description: "Query type"
						validations: ["value in ['select', 'insert', 'update', 'delete', 'upsert']"]
					}
					table: {
						description: "Target table name"
						validations: [
							"value.matches('^[a-z_][a-z0-9_]*$')",
							"size(value) <= 63",
						]
					}
					status: {
						description: "Query result status"
						validations: ["value in ['success', 'error', 'timeout', 'cancelled']"]
					}
				}
				examples: {
					promql: [
						{
							query:       "sum(rate(db_postgres_queries_total[5m])) by (operation)"
							description: "Query rate by operation type"
						},
						{
							query:       "sum(rate(db_postgres_queries_total{status=\"error\"}[5m])) by (table)"
							description: "Error rate by table"
						},
					]
					alerts: [
						{
							name:        "DatabaseHighErrorRate"
							expr:        "sum(rate(db_postgres_queries_total{status=\"error\"}[5m])) / sum(rate(db_postgres_queries_total[5m])) > 0.001"
							description: "Database error rate exceeds 0.1%"
							for:         "5m"
							severity:    "critical"
						},
					]
				}
			}

			db_query_duration_seconds: {
				namespace: "db"
				subsystem: "postgres"
				type:      "histogram"
				help:      "Database query duration in seconds"
				buckets: [0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5]
				labels: {
					operation: {description: "Query type (select, insert, update, delete)"}
					table:     {description: "Target table name"}
				}
				examples: {
					promql: [
						{
							query:       "histogram_quantile(0.95, sum(rate(db_postgres_query_duration_seconds_bucket[5m])) by (le, table))"
							description: "P95 query latency by table"
						},
					]
					alerts: [
						{
							name:        "SlowQueries"
							expr:        "histogram_quantile(0.95, rate(db_postgres_query_duration_seconds_bucket[5m])) > 0.1"
							description: "P95 query latency exceeds 100ms"
							for:         "10m"
							severity:    "warning"
						},
					]
				}
			}

			db_connections_active: {
				namespace: "db"
				subsystem: "postgres"
				type:      "gauge"
				help:      "Number of active database connections in the pool"
				examples: {
					promql: [
						{
							query:       "db_postgres_connections_active / db_postgres_connections_max"
							description: "Connection pool utilization"
						},
					]
				}
			}

			db_connections_max: {
				namespace: "db"
				subsystem: "postgres"
				type:      "gauge"
				help:      "Maximum number of connections allowed in the pool"
			}

			db_connections_idle: {
				namespace: "db"
				subsystem: "postgres"
				type:      "gauge"
				help:      "Number of idle connections in the pool"
			}

			db_connections_wait_seconds: {
				namespace: "db"
				subsystem: "postgres"
				type:      "histogram"
				help:      "Time spent waiting for a database connection from the pool"
				buckets: [0.0001, 0.001, 0.01, 0.1, 0.5, 1, 5]
				examples: {
					alerts: [
						{
							name:        "ConnectionPoolExhausted"
							expr:        "histogram_quantile(0.95, rate(db_postgres_connections_wait_seconds_bucket[5m])) > 0.1"
							description: "Waiting too long for database connections"
							for:         "5m"
							severity:    "critical"
						},
					]
				}
			}

			// =========================================
			// Cache metrics
			// =========================================
			cache_requests_total: {
				namespace: "cache"
				subsystem: "redis"
				type:      "counter"
				help:      "Total number of cache requests"
				labels: {
					operation: {
						description: "Cache operation type"
						validations: ["value in ['get', 'set', 'del', 'mget', 'mset', 'expire', 'exists']"]
					}
					status: {
						description: "Operation result"
						validations: ["value in ['hit', 'miss', 'error']"]
					}
				}
				examples: {
					promql: [
						{
							query:       "sum(rate(cache_redis_requests_total{status=\"hit\"}[5m])) / sum(rate(cache_redis_requests_total{operation=\"get\"}[5m]))"
							description: "Cache hit ratio"
						},
						{
							query:       "sum(rate(cache_redis_requests_total{status=\"miss\"}[5m]))"
							description: "Cache miss rate"
						},
					]
					alerts: [
						{
							name:        "LowCacheHitRatio"
							expr:        "sum(rate(cache_redis_requests_total{status=\"hit\"}[5m])) / sum(rate(cache_redis_requests_total{operation=\"get\"}[5m])) < 0.8"
							description: "Cache hit ratio below 80%"
							for:         "15m"
							severity:    "warning"
						},
					]
				}
			}

			cache_operation_duration_seconds: {
				namespace: "cache"
				subsystem: "redis"
				type:      "histogram"
				help:      "Cache operation duration in seconds"
				buckets: [0.00001, 0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1]
				labels: {
					operation: {description: "Cache operation type"}
				}
			}

			cache_keys_total: {
				namespace: "cache"
				subsystem: "redis"
				type:      "gauge"
				help:      "Total number of keys in the cache"
			}

			cache_memory_bytes: {
				namespace: "cache"
				subsystem: "redis"
				type:      "gauge"
				help:      "Memory used by the cache in bytes"
				examples: {
					alerts: [
						{
							name:        "CacheMemoryHigh"
							expr:        "cache_redis_memory_bytes > 1e9"
							description: "Cache memory usage exceeds 1GB"
							for:         "10m"
							severity:    "warning"
						},
					]
				}
			}

			// =========================================
			// Business metrics
			// =========================================
			orders_created_total: {
				namespace: "business"
				subsystem: "orders"
				type:      "counter"
				help:      "Total number of orders created"
				labels: {
					payment_method: {
						description: "Payment method used"
						validations: ["value in ['credit_card', 'debit_card', 'paypal', 'apple_pay', 'google_pay', 'bank_transfer']"]
					}
					status: {
						description: "Order creation status"
						validations: ["value in ['success', 'failed', 'pending']"]
					}
					channel: {
						description: "Sales channel"
						validations: ["value in ['web', 'mobile', 'api', 'pos']"]
					}
				}
				examples: {
					promql: [
						{
							query:       "sum(rate(business_orders_created_total{status=\"success\"}[1h])) * 3600"
							description: "Orders per hour"
						},
						{
							query:       "sum(rate(business_orders_created_total[5m])) by (channel)"
							description: "Order rate by sales channel"
						},
					]
				}
			}

			orders_value_total: {
				namespace: "business"
				subsystem: "orders"
				type:      "counter"
				help:      "Total monetary value of orders (in cents)"
				labels: {
					currency: {
						description: "Order currency (ISO 4217)"
						validations: ["value.matches('^[A-Z]{3}$')"]
					}
					status: {
						description: "Order status"
					}
				}
				examples: {
					promql: [
						{
							query:       "sum(rate(business_orders_value_total{currency=\"USD\",status=\"success\"}[1h])) * 3600 / 100"
							description: "Revenue per hour in USD"
						},
					]
				}
			}

			orders_processing_duration_seconds: {
				namespace: "business"
				subsystem: "orders"
				type:      "histogram"
				help:      "Time taken to process an order from creation to completion"
				buckets: [0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300]
				labels: {
					payment_method: {description: "Payment method used"}
				}
			}

			// =========================================
			// Deprecated metric example
			// =========================================
			http_request_count: {
				namespace: "http"
				subsystem: "server"
				type:      "counter"
				help:      "Total HTTP request count"
				labels: {
					method: {description: "HTTP method"}
					code:   {description: "HTTP status code"}
				}
				 deprecated: {
				 	since:      "1.2.0"
				 	replacedBy: "http_requests_total"
				 	reason:     "Renamed for consistency with Prometheus naming conventions"
				 }
			}
		}

		goldenSignals: {
			// =========================================
			// HTTP Golden Signals
			// =========================================
			"http/server": {
				latency: {
					description: "How long HTTP requests take to complete. Key indicator of user experience."
					metrics: ["http_request_duration_seconds"]
					recordingRules: [
						{
							name:  "http:server:latency:p50:5m"
							query: "histogram_quantile(0.50, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
						},
						{
							name:  "http:server:latency:p95:5m"
							query: "histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
						},
						{
							name:  "http:server:latency:p99:5m"
							query: "histogram_quantile(0.99, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le))"
						},
						{
							name:  "http:server:latency:p95:5m:by_path"
							query: "histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le, path))"
						},
					]
					thresholds: {
						good:     "< 100ms"
						warning:  "< 500ms"
						critical: ">= 500ms"
					}
				}
				errors: {
					description: "Rate of failed HTTP requests. Includes 5xx server errors and application errors."
					metrics: ["http_requests_total"]
					recordingRules: [
						{
							name:  "http:server:errors:ratio:5m"
							query: "sum(rate(http_server_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_server_requests_total[5m]))"
						},
						{
							name:  "http:server:errors:rate:5m"
							query: "sum(rate(http_server_requests_total{status=~\"5..\"}[5m]))"
						},
						{
							name:  "http:server:errors:ratio:5m:by_path"
							query: "sum(rate(http_server_requests_total{status=~\"5..\"}[5m])) by (path) / sum(rate(http_server_requests_total[5m])) by (path)"
						},
					]
					thresholds: {
						good:     "< 0.1%"
						warning:  "< 1%"
						critical: ">= 1%"
					}
				}
				traffic: {
					description: "Request volume handled by the service. Useful for capacity planning and anomaly detection."
					metrics: ["http_requests_total"]
					recordingRules: [
						{
							name:  "http:server:traffic:rate:5m"
							query: "sum(rate(http_server_requests_total[5m]))"
						},
						{
							name:  "http:server:traffic:rate:5m:by_method"
							query: "sum(rate(http_server_requests_total[5m])) by (method)"
						},
						{
							name:  "http:server:traffic:rate:5m:by_path"
							query: "sum(rate(http_server_requests_total[5m])) by (path)"
						},
					]
				}
				saturation: {
					description: "How close to capacity the HTTP server is. High saturation indicates potential bottlenecks."
					metrics: ["http_requests_in_flight"]
					recordingRules: [
						{
							name:  "http:server:saturation:in_flight"
							query: "sum(http_server_requests_in_flight)"
						},
						{
							name:  "http:server:saturation:in_flight:max_1h"
							query: "max_over_time(sum(http_server_requests_in_flight)[1h:])"
						},
					]
					thresholds: {
						good:     "< 100"
						warning:  "< 500"
						critical: ">= 500"
					}
				}
			}

			// =========================================
			// Database Golden Signals
			// =========================================
			"db/postgres": {
				latency: {
					description: "Database query execution time. Slow queries directly impact application performance."
					metrics: ["db_query_duration_seconds"]
					recordingRules: [
						{
							name:  "db:postgres:latency:p50:5m"
							query: "histogram_quantile(0.50, sum(rate(db_postgres_query_duration_seconds_bucket[5m])) by (le))"
						},
						{
							name:  "db:postgres:latency:p95:5m"
							query: "histogram_quantile(0.95, sum(rate(db_postgres_query_duration_seconds_bucket[5m])) by (le))"
						},
						{
							name:  "db:postgres:latency:p99:5m"
							query: "histogram_quantile(0.99, sum(rate(db_postgres_query_duration_seconds_bucket[5m])) by (le))"
						},
						{
							name:  "db:postgres:latency:p95:5m:by_table"
							query: "histogram_quantile(0.95, sum(rate(db_postgres_query_duration_seconds_bucket[5m])) by (le, table))"
						},
						{
							name:  "db:postgres:latency:p95:5m:by_operation"
							query: "histogram_quantile(0.95, sum(rate(db_postgres_query_duration_seconds_bucket[5m])) by (le, operation))"
						},
					]
					thresholds: {
						good:     "< 10ms"
						warning:  "< 50ms"
						critical: ">= 50ms"
					}
				}
				errors: {
					description: "Database query failures. Even low error rates can indicate serious issues."
					metrics: ["db_queries_total"]
					recordingRules: [
						{
							name:  "db:postgres:errors:ratio:5m"
							query: "sum(rate(db_postgres_queries_total{status=\"error\"}[5m])) / sum(rate(db_postgres_queries_total[5m]))"
						},
						{
							name:  "db:postgres:errors:rate:5m"
							query: "sum(rate(db_postgres_queries_total{status=\"error\"}[5m]))"
						},
						{
							name:  "db:postgres:errors:ratio:5m:by_table"
							query: "sum(rate(db_postgres_queries_total{status=\"error\"}[5m])) by (table) / sum(rate(db_postgres_queries_total[5m])) by (table)"
						},
					]
					thresholds: {
						good:     "< 0.01%"
						critical: ">= 0.01%"
					}
				}
				traffic: {
					description: "Query volume to the database. Helps identify hot tables and capacity needs."
					metrics: ["db_queries_total"]
					recordingRules: [
						{
							name:  "db:postgres:traffic:rate:5m"
							query: "sum(rate(db_postgres_queries_total[5m]))"
						},
						{
							name:  "db:postgres:traffic:rate:5m:by_operation"
							query: "sum(rate(db_postgres_queries_total[5m])) by (operation)"
						},
						{
							name:  "db:postgres:traffic:rate:5m:by_table"
							query: "sum(rate(db_postgres_queries_total[5m])) by (table)"
						},
					]
				}
				saturation: {
					description: "Connection pool utilization. Exhausted pools cause request queuing and timeouts."
					metrics: ["db_connections_active", "db_connections_max"]
					recordingRules: [
						{
							name:  "db:postgres:saturation:pool_utilization"
							query: "db_postgres_connections_active / db_postgres_connections_max"
						},
						{
							name:  "db:postgres:saturation:pool_utilization:max_1h"
							query: "max_over_time((db_postgres_connections_active / db_postgres_connections_max)[1h:])"
						},
						{
							name:  "db:postgres:saturation:connections_available"
							query: "db_postgres_connections_max - db_postgres_connections_active"
						},
					]
					thresholds: {
						good:     "< 70%"
						warning:  "< 90%"
						critical: ">= 90%"
					}
				}
			}

			// =========================================
			// Cache Golden Signals
			// =========================================
			"cache/redis": {
				latency: {
					description: "Cache operation latency. Should be sub-millisecond for optimal performance."
					metrics: ["cache_operation_duration_seconds"]
					recordingRules: [
						{
							name:  "cache:redis:latency:p50:5m"
							query: "histogram_quantile(0.50, sum(rate(cache_redis_operation_duration_seconds_bucket[5m])) by (le))"
						},
						{
							name:  "cache:redis:latency:p95:5m"
							query: "histogram_quantile(0.95, sum(rate(cache_redis_operation_duration_seconds_bucket[5m])) by (le))"
						},
						{
							name:  "cache:redis:latency:p99:5m"
							query: "histogram_quantile(0.99, sum(rate(cache_redis_operation_duration_seconds_bucket[5m])) by (le))"
						},
					]
					thresholds: {
						good:     "< 1ms"
						warning:  "< 5ms"
						critical: ">= 5ms"
					}
				}
				errors: {
					description: "Cache errors and miss rate. High miss rates impact database load and latency."
					metrics: ["cache_requests_total"]
					recordingRules: [
						{
							name:  "cache:redis:hit_ratio:5m"
							query: "sum(rate(cache_redis_requests_total{status=\"hit\"}[5m])) / sum(rate(cache_redis_requests_total{operation=\"get\"}[5m]))"
						},
						{
							name:  "cache:redis:miss_ratio:5m"
							query: "sum(rate(cache_redis_requests_total{status=\"miss\"}[5m])) / sum(rate(cache_redis_requests_total{operation=\"get\"}[5m]))"
						},
						{
							name:  "cache:redis:error_ratio:5m"
							query: "sum(rate(cache_redis_requests_total{status=\"error\"}[5m])) / sum(rate(cache_redis_requests_total[5m]))"
						},
					]
					thresholds: {
						good:     "< 10% miss"
						warning:  "< 30% miss"
						critical: ">= 30% miss"
					}
				}
				traffic: {
					description: "Cache operations per second. High traffic indicates heavy cache usage."
					metrics: ["cache_requests_total"]
					recordingRules: [
						{
							name:  "cache:redis:traffic:rate:5m"
							query: "sum(rate(cache_redis_requests_total[5m]))"
						},
						{
							name:  "cache:redis:traffic:rate:5m:by_operation"
							query: "sum(rate(cache_redis_requests_total[5m])) by (operation)"
						},
					]
				}
				saturation: {
					description: "Cache memory utilization and key count. Indicates when cache eviction may occur."
					metrics: ["cache_memory_bytes", "cache_keys_total"]
					recordingRules: [
						{
							name:  "cache:redis:saturation:memory_bytes"
							query: "cache_redis_memory_bytes"
						},
						{
							name:  "cache:redis:saturation:keys_total"
							query: "cache_redis_keys_total"
						},
					]
					thresholds: {
						good:     "< 70% memory"
						warning:  "< 90% memory"
						critical: ">= 90% memory"
					}
				}
			}
		}
	}
}
