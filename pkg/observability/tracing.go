package observability

// Tracing plan (future implementation)
//
// Goal:
// - Provide a minimal OpenTelemetry wrapper for distributed tracing.
// - Keep API simple for handlers/services/repositories.
//
// Proposed scope:
// 1) Initialization
//    - Add InitTracing(config) to configure TracerProvider.
//    - Support service name and exporter setup (OTEL collector, Jaeger, or stdout for local debug).
//    - Return a shutdown function to flush and close exporter.
//
// 2) Span API
//    - Add StartSpan(ctx, tracerName, spanName, opts...) helper.
//    - Return updated context plus end function to close span.
//    - Record error when end function receives a non-nil error.
//
// 3) Usage pattern
//    - Start a span in API handlers and pass context through all internal calls.
//    - Add child spans only around meaningful boundaries:
//      DB operations, blockchain provider calls, queue publish/consume.
//
// 4) Safety and defaults
//    - If tracing is not configured, keep a no-op behavior (no crashes).
//    - Avoid high-cardinality span attributes (wallet, tx hash, user id).
