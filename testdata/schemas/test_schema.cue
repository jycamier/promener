package testschema

// Simple test schema for validation
#Specification: {
	version: string & =~"^[0-9]+\\.[0-9]+(\\.[0-9]+)?$"
	info: {
		title:       string
		description?: string
		version:     string
	}
	services: [string]: #Service
}

#Service: {
	info: {
		title:       string
		description?: string
		version:     string
	}
	servers?: [...#Server]
	metrics: [string]: #Metric
}

#Server: {
	url:          string
	description?: string
}

#Metric: {
	name?:      string
	namespace:  string
	subsystem:  string
	type:       "counter" | "gauge" | "histogram" | "summary"
	help:       string
	labels?:    [string]: #Label | string
	constLabels?: [string]: #ConstLabel
	buckets?:   [...number]
}

#Label: {
	description: string
}

#ConstLabel: {
	value: string
}
