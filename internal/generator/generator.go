package generator

// Generator is a deprecated alias for GoGenerator
// Maintained for backward compatibility
type Generator = GoGenerator

// New creates a new Generator instance (backward compatibility)
// Use NewGoGenerator() for new code
func New() (*Generator, error) {
	return NewGoGenerator()
}

