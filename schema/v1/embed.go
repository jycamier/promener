package v1

import _ "embed"

// Schema contains the embedded CUE schema for v1 specifications.
//
//go:embed schema.cue
var Schema string
