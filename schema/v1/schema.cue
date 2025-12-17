package v1

// =============================================================================
// Common definitions
// =============================================================================

#Info: {
	title:        string
	description?: string
	version:      string
	package?:     string
}

#Server: {
	url:         string
	description: string
}

// =============================================================================
// Metrics definitions
// =============================================================================

#PromQLExample: {
	query:       string
	description: string
}

#AlertExample: {
	name:        string
	expr:        string
	description: string
	for:         string
	severity:    "info" | "warning" | "critical"
	labels?: [string]:      string
	annotations?: [string]: string
}

#Metric: {
	namespace:  string
	subsystem?: string
	type:       "counter" | "gauge" | "histogram" | "summary"
	help:       string
	labels?: [string]: {
		description: string
		validations?: [...string]
		inherited?: string
	}
	constLabels?: [string]: {
		value:       string
		description: string
	}
	buckets?: [...number]
	objectives?: [string]: number
	examples?: {
		promql?: [...#PromQLExample]
		alerts?: [...#AlertExample]
	}
}

// =============================================================================
// Golden Signals definitions
// =============================================================================

#RecordingRule: {
	// Name of the recording rule (will be used as metric name)
	name: string

	// PromQL query
	query: string
}

#Thresholds: {
	// Value considered good (green)
	good: string

	// Value considered warning (yellow)
	warning?: string

	// Value considered critical (red)
	critical: string
}

#GoldenSignal: {
	// What this signal measures
	description: string

	// References to metric names defined in this service
	metrics: [...string]

	// Pre-computed recording rules for dashboards
	recordingRules?: [...#RecordingRule]

	// Thresholds for dashboard visualization
	thresholds?: #Thresholds
}

#GoldenSignals: {
	// Latency: How long it takes to service a request
	latency?: #GoldenSignal

	// Errors: The rate of requests that fail
	errors?: #GoldenSignal

	// Traffic: How much demand is being placed on your system
	traffic?: #GoldenSignal

	// Saturation: How "full" your service is
	saturation?: #GoldenSignal
}

// =============================================================================
// Root schema
// =============================================================================

#Promener: {
	version: string | *"1.0"
	info:    #Info
	services?: [string]: {
		info:    #Info
		servers?: [...#Server]
		metrics: [string]: #Metric

		// Golden Signals by topic (http, database, cache, queue, etc.)
		goldenSignals?: [string]: #GoldenSignals
	}
}
