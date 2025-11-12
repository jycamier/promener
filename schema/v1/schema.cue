package v1

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

#Promener: {
	version: string | *"1.0"
	info:    #Info
	services?: [string]: {
		info: #Info
		servers?: [...#Server]
		metrics: [string]: #Metric
	}
}
