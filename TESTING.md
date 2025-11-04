# Testing Guide

This document describes how to run and write tests for promener.

## Quick Start

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Generate HTML coverage report
make coverage-html
# Then open coverage.html in your browser

# Run tests for a specific package
go test ./internal/validator -v

# Run a specific test
go test ./internal/validator -v -run TestCueLoader_LoadAndValidate
```

### Generating Mocks

Mocks are generated using `mockgen` from interfaces defined in the codebase.

```bash
# Generate all mocks
make generate

# Or manually with go generate
go generate ./...
```

## Test Coverage

Current coverage status:

| Package | Coverage |
|---------|----------|
| internal/validator | 75.5% |
| internal/generator | 54.6% |
| internal/domain | 45.0% |
| **Total** | **41.2%** |

## Test Structure

### Unit Tests

Unit tests are located alongside the code they test:

```
internal/
├── validator/
│   ├── validator.go
│   ├── validator_test.go          # Unit tests
│   ├── cue_loader.go
│   ├── cue_loader_test.go
│   ├── interface.go                # Interfaces for mocking
│   └── mocks/
│       └── mock_validator.go       # Generated mocks
```

### Test Categories

1. **Domain Tests** (`internal/domain/*_test.go`)
   - Metric validation
   - Label validation
   - CEL expression parsing and validation

2. **Validator Tests** (`internal/validator/*_test.go`)
   - CUE file loading and validation
   - Schema version handling
   - Spec extraction from CUE
   - Formatter output (text/JSON)

3. **Generator Tests** (`internal/generator/*_test.go`)
   - Go code generation
   - Template data building
   - Environment variable parsing

## Writing Tests

### Using Table-Driven Tests

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "TEST",
            wantErr: false,
        },
        {
            name:    "empty input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Using Mocks

1. **Define an interface** in your package (e.g., `interface.go`):

```go
//go:generate mockgen -source=interface.go -destination=mocks/mock_myinterface.go -package=mocks

package mypackage

type MyInterface interface {
    DoSomething(input string) (string, error)
}
```

2. **Generate mocks**:

```bash
go generate ./...
```

3. **Use mocks in tests**:

```go
import (
    "testing"
    "github.com/jycamier/promener/internal/mypackage/mocks"
    "go.uber.org/mock/gomock"
)

func TestWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockObj := mocks.NewMockMyInterface(ctrl)
    mockObj.EXPECT().
        DoSomething("input").
        Return("output", nil).
        Times(1)

    // Use mockObj in your test
    result, err := mockObj.DoSomething("input")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != "output" {
        t.Errorf("got %q, want %q", result, "output")
    }
}
```

### Testing with Temporary Files

```go
func TestWithTempFile(t *testing.T) {
    // Create temp directory
    tmpDir := t.TempDir() // Automatically cleaned up

    // Create temp file
    tmpFile := filepath.Join(tmpDir, "test.cue")
    content := `version: "1.0.0"`

    err := os.WriteFile(tmpFile, []byte(content), 0644)
    if err != nil {
        t.Fatalf("failed to create temp file: %v", err)
    }

    // Use tmpFile in your test
    // ...
}
```

## CI/CD Integration

Tests run automatically on:
- Every push to `main` branch
- Every pull request

The GitHub Actions workflow (`.github/workflows/test.yaml`) runs:
1. All tests with race detection
2. Coverage report generation
3. Coverage upload to Codecov
4. Linting with golangci-lint

## Best Practices

1. **Test Names**: Use descriptive names that explain what is being tested
   - Good: `TestCueLoader_LoadAndValidate_WithValidFile`
   - Bad: `TestLoadFile`

2. **Test One Thing**: Each test should focus on one specific behavior

3. **Use Subtests**: Group related test cases with `t.Run()`

4. **Mock External Dependencies**: Don't make real HTTP calls, file I/O (when possible), etc.

5. **Clean Up**: Always clean up resources (files, connections) in `defer` or `t.Cleanup()`

6. **Coverage Goal**: Aim for >70% coverage for critical packages (validator, generator, domain)

## Troubleshooting

### Mocks not found

```bash
# Regenerate all mocks
make generate
```

### Tests failing on CI but passing locally

- Check Go version matches (1.25.1)
- Ensure `go generate` was run
- Check for race conditions: `go test ./... -race`

### Coverage not updating

```bash
# Clear cache and regenerate
go clean -testcache
make coverage
```

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [gomock Documentation](https://github.com/uber-go/mock)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Test Coverage Best Practices](https://go.dev/blog/cover)
