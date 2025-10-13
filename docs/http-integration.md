# HTTP Server Integration

This guide shows you how to integrate Promener-generated metrics with your HTTP server.

## Table of Contents

- [Using the Default Registry](#using-the-default-registry)
- [Using a Custom Registry](#using-a-custom-registry)
- [Complete Example](#complete-example)
- [Middleware Pattern](#middleware-pattern)
- [Best Practices](#best-practices)

## Using the Default Registry

The simplest way to use metrics is with the default Prometheus registry:

```go
package main

import (
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus/promhttp"
    "yourapp/metrics"
)

func main() {
    // Initialize metrics with default registry
    m := metrics.Default()

    // Expose metrics on /metrics endpoint
    http.Handle("/metrics", promhttp.Handler())

    // Your application routes
    http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Your business logic here
        w.Write([]byte("OK"))

        // Record metrics
        duration := time.Since(start).Seconds()
        m.Http.Server.IncRequestsTotal(r.Method, "200", "/api/users")
        m.Http.Server.ObserveRequestDurationSeconds(r.Method, "/api/users", duration)
    })

    http.ListenAndServe(":8080", nil)
}
```

### Why use the default registry?

- Simple and straightforward
- Works with `promhttp.Handler()` out of the box
- Includes default Go runtime metrics (goroutines, GC, memory, etc.)

## Using a Custom Registry

For better isolation and control, use a custom registry:

```go
package main

import (
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "yourapp/metrics"
)

func main() {
    // Create your own registry
    registry := prometheus.NewRegistry()

    // Initialize metrics with custom registry
    m := metrics.NewMetricsRegistry(registry)

    // Create handler for your custom registry
    http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

    // Your application routes
    http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Your business logic here
        w.Write([]byte("OK"))

        // Record metrics
        duration := time.Since(start).Seconds()
        m.Http.Server.IncRequestsTotal(r.Method, "200", "/api/users")
        m.Http.Server.ObserveRequestDurationSeconds(r.Method, "/api/users", duration)
    })

    http.ListenAndServe(":8080", nil)
}
```

### Why use a custom registry?

- **Isolation**: Only your application metrics, no Go runtime metrics
- **Control**: Full control over what gets exposed
- **Testing**: Easier to test in isolation
- **Multiple registries**: Can have different metric sets for different purposes

### Adding Go Runtime Metrics

If you want Go runtime metrics with your custom registry:

```go
registry := prometheus.NewRegistry()

// Add Go collector (goroutines, GC, memory, etc.)
registry.MustRegister(prometheus.NewGoCollector())

// Add process collector (CPU, file descriptors, etc.)
registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

// Initialize your metrics
m := metrics.NewMetricsRegistry(registry)
```

## Complete Example

Here's a complete example with error handling and middleware:

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "yourapp/metrics"
)

func main() {
    // Setup registry
    registry := prometheus.NewRegistry()
    registry.MustRegister(prometheus.NewGoCollector())
    registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

    // Initialize metrics
    m := metrics.NewMetricsRegistry(registry)

    // Setup HTTP server
    mux := http.NewServeMux()

    // Metrics endpoint
    mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
        ErrorHandling: promhttp.ContinueOnError,
    }))

    // Application routes with metrics middleware
    mux.Handle("/api/users", metricsMiddleware(m, http.HandlerFunc(usersHandler)))
    mux.Handle("/api/posts", metricsMiddleware(m, http.HandlerFunc(postsHandler)))

    // Health check (no metrics)
    mux.HandleFunc("/health", healthHandler)

    // Create server
    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    // Graceful shutdown
    go func() {
        log.Println("Server starting on :8080")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exited")
}

func metricsMiddleware(m *metrics.MetricsRegistry, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Increment active connections
        m.Http.Server.IncActiveConnections(r.Proto)
        defer m.Http.Server.DecActiveConnections(r.Proto)

        // Wrap response writer to capture status code
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        // Call next handler
        next.ServeHTTP(rw, r)

        // Record metrics
        duration := time.Since(start).Seconds()
        statusStr := http.StatusText(rw.statusCode)

        m.Http.Server.IncRequestsTotal(r.Method, statusStr, r.URL.Path)
        m.Http.Server.ObserveRequestDurationSeconds(r.Method, r.URL.Path, duration)
    })
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Users endpoint"))
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Posts endpoint"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("OK"))
}
```

## Middleware Pattern

For cleaner code, create a reusable middleware:

```go
type MetricsMiddleware struct {
    metrics *metrics.MetricsRegistry
}

func NewMetricsMiddleware(m *metrics.MetricsRegistry) *MetricsMiddleware {
    return &MetricsMiddleware{metrics: m}
}

func (mm *MetricsMiddleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Track active connections
        mm.metrics.Http.Server.IncActiveConnections(r.Proto)
        defer mm.metrics.Http.Server.DecActiveConnections(r.Proto)

        // Wrap response writer
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        // Process request
        next.ServeHTTP(rw, r)

        // Record metrics
        duration := time.Since(start).Seconds()
        mm.metrics.Http.Server.IncRequestsTotal(r.Method,
            http.StatusText(rw.statusCode), r.URL.Path)
        mm.metrics.Http.Server.ObserveRequestDurationSeconds(r.Method,
            r.URL.Path, duration)
    })
}

// Usage
func main() {
    registry := prometheus.NewRegistry()
    m := metrics.NewMetricsRegistry(registry)
    metricsMiddleware := NewMetricsMiddleware(m)

    mux := http.NewServeMux()
    mux.Handle("/api/", metricsMiddleware.Handler(http.HandlerFunc(apiHandler)))

    http.ListenAndServe(":8080", mux)
}
```

## Best Practices

### 1. Initialize Once

Metrics are initialized using `sync.Once`, so it's safe to call `NewMetricsRegistry()` multiple times, but typically you should initialize once at startup:

```go
var globalMetrics *metrics.MetricsRegistry

func init() {
    registry := prometheus.NewRegistry()
    globalMetrics = metrics.NewMetricsRegistry(registry)
}
```

### 2. Use Dependency Injection

Pass the metrics registry as a dependency:

```go
type Server struct {
    metrics *metrics.MetricsRegistry
}

func NewServer(m *metrics.MetricsRegistry) *Server {
    return &Server{metrics: m}
}

func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
    s.metrics.Http.Server.IncRequestsTotal(r.Method, "200", r.URL.Path)
}
```

### 3. Label Cardinality

Be careful with high-cardinality labels (like user IDs or request IDs). This can cause memory issues:

```go
// ❌ BAD: user_id has high cardinality
m.Http.Server.IncRequestsTotal(r.Method, "200", userID)

// ✅ GOOD: Use bounded labels
m.Http.Server.IncRequestsTotal(r.Method, "200", "/api/users")
```

### 4. Error Handling

Always handle errors when recording metrics:

```go
defer func() {
    if err := recover(); err != nil {
        m.Http.Server.IncRequestsTotal(r.Method, "500", r.URL.Path)
        panic(err)
    }
}()
```

### 5. Testing

For testing, use a custom registry:

```go
func TestHandler(t *testing.T) {
    registry := prometheus.NewRegistry()
    m := metrics.NewMetricsRegistry(registry)

    // Test your handler
    handler := NewHandler(m)

    // Verify metrics were recorded
    metricFamilies, _ := registry.Gather()
    // Assert on metrics
}
```

### 6. Metrics Endpoint Security

In production, consider protecting the metrics endpoint:

```go
http.Handle("/metrics", basicAuth(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
```

## Troubleshooting

### Metrics not showing up

1. Check that the metrics endpoint is accessible: `curl http://localhost:8080/metrics`
2. Verify metrics are being recorded in your code
3. Check that the registry is properly initialized

### Duplicate metrics error

If you see "duplicate metrics collector registration":
- Ensure you're using `sync.Once` (already handled by generated code)
- Don't create multiple registries with the same metrics

### High memory usage

- Check for high-cardinality labels
- Use bounded label values
- Consider using summaries instead of histograms for high-volume metrics

## Next Steps

- [YAML Specification](yaml-specification.md) - Learn about the YAML format
- [Generated Code Structure](generated-code.md) - Understand the generated code
