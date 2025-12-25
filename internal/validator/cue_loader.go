package validator

import (
	"fmt"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
)

// CueLoader loads and validates CUE files against embedded schemas.
type CueLoader struct {
	ctx *cue.Context
}

// NewCueLoader creates a new CUE loader.
func NewCueLoader() *CueLoader {
	return &CueLoader{
		ctx: cuecontext.New(),
	}
}

// LoadAndValidate loads a CUE file, determines its version, and validates it against the embedded schema.
// Returns the validated CUE value and any validation errors.
func (l *CueLoader) LoadAndValidate(cuePath string) (cue.Value, *ValidationResult, error) {
	result := &ValidationResult{
		CueErrors:    []ValidationError{},
		DomainErrors: []ValidationError{},
		RegoErrors:   []ValidationError{},
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(cuePath)
	if err != nil {
		return cue.Value{}, nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get the directory containing the CUE file
	// This is needed so cue/load can find cue.mod in the file's directory
	fileDir := filepath.Dir(absPath)

	// Use cue/load to load the file with module context
	// Set Dir to the file's directory so it can find cue.mod
	config := &load.Config{
		Dir: fileDir,
	}
	instances := load.Instances([]string{filepath.Base(absPath)}, config)
	if len(instances) == 0 {
		return cue.Value{}, nil, fmt.Errorf("no instances loaded from %s", absPath)
	}

	// Check for load errors
	inst := instances[0]
	if inst.Err != nil {
		return cue.Value{}, nil, fmt.Errorf("failed to load CUE file: %w", inst.Err)
	}

	// Build the instance into a value
	specValue := l.ctx.BuildInstance(inst)
	if specValue.Err() != nil {
		return cue.Value{}, nil, fmt.Errorf("failed to build CUE instance: %w", specValue.Err())
	}

	// Extract the version field
	versionValue := specValue.LookupPath(cue.ParsePath("version"))
	if !versionValue.Exists() {
		return cue.Value{}, nil, fmt.Errorf("'version' field is required in the CUE specification")
	}

	version, err := versionValue.String()
	if err != nil {
		return cue.Value{}, nil, fmt.Errorf("'version' field must be a string: %w", err)
	}

	// Get the schema for this version
	schemaContent, err := GetSchemaForVersion(version)
	if err != nil {
		return cue.Value{}, nil, err
	}

	// Compile the schema
	schemaValue := l.ctx.CompileString(schemaContent, cue.Scope(specValue))
	if schemaValue.Err() != nil {
		return cue.Value{}, nil, fmt.Errorf("failed to compile embedded schema: %w", schemaValue.Err())
	}

	// Unify the specification with the schema
	unified := schemaValue.Unify(specValue)

	// Validate
	if err := unified.Validate(cue.Concrete(true), cue.All()); err != nil {
		cueErrors := l.extractCueErrors(err)
		result.CueErrors = append(result.CueErrors, cueErrors...)
	}

	// Return the validated value
	return specValue, result, nil
}

// extractCueErrors converts CUE errors into ValidationError structs.
func (l *CueLoader) extractCueErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	// CUE errors can be a list of errors
	for _, e := range errors.Errors(err) {
		pos := e.Position()
		path := l.extractPath(e)

		validationErrors = append(validationErrors, ValidationError{
			Path:     path,
			Message:  strings.TrimSpace(e.Error()),
			Source:   "cue",
			Severity: "error",
			Line:     pos.Line(),
		})
	}

	return validationErrors
}

// extractPath attempts to extract the field path from a CUE error.
func (l *CueLoader) extractPath(err errors.Error) string {
	// Try to get the path from the error position
	pos := err.Position()
	if pos.IsValid() {
		// The error message often contains the path
		msg := err.Error()
		// Extract path pattern like "services.default.metrics.test_test"
		if idx := strings.Index(msg, ":"); idx > 0 {
			pathPart := msg[:idx]
			pathPart = strings.TrimSpace(pathPart)
			// Remove line/column info if present
			if spaceIdx := strings.Index(pathPart, " "); spaceIdx > 0 {
				pathPart = pathPart[:spaceIdx]
			}
			return pathPart
		}
	}
	return ""
}
