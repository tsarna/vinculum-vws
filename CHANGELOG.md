# Changelog

## [Unreleased]

## [0.11.0] - 2026-04-08

### Changed

- **OTel metrics replaces o11y.MetricsProvider abstraction** — the listener now accepts `metric.MeterProvider` directly via `WithMeterProvider()` (replacing `WithMetricsProvider(o11y.MetricsProvider)`). Metric names follow OTel naming conventions with dot-delimited hierarchy: `websocket.connections`, `websocket.active_connections`, `websocket.connection.duration`, `websocket.received.messages`, `websocket.sent.messages`, `websocket.message.size`, `websocket.requests`, `websocket.request.duration`, `websocket.pings_sent`, `websocket.pong_timeouts`, `websocket.write_timeouts`, etc. Label keys updated to OTel conventions (`error.type` instead of `error_type`, `websocket.message.kind` instead of `kind`). Requires vinculum-bus v0.11.0.

## [0.10.0] - 2026-04-03

### Added

- **Distributed tracing support**: VWS wire messages now carry a generic headers field (`"h"`) as a `map[string]string`, used to propagate W3C TraceContext (`traceparent`, `tracestate`) and Baggage across the WebSocket boundary.
- New package-level helpers in the root `vws` package: `InjectTrace`, `ExtractTrace`, and `HeadersFromContext`, which use the OpenTelemetry global `TextMapPropagator`. No tracer provider configuration is required in vws itself — applications configure the propagator once globally (e.g. with `propagation.TraceContext{}`).
- Server: incoming message trace context is extracted and threaded through to EventBus publish calls; outbound events, ACK, and NACK responses carry the active trace context.
- Client: all outgoing messages (publish, subscribe, unsubscribe) carry the caller's trace context; incoming events restore trace context before invoking the subscriber's `OnEvent`.
- Updated `PROTOCOL.md` to document the new `"h"` field and well-known header keys.

### Changed

- `WireMessage` gains a new `Headers map[string]string` field (`json:"h,omitempty"`). The field is omitted when empty, so existing clients and servers that do not set it remain fully compatible.
- Updated `vinculum-bus` dependency to v0.10.0.
- Added `go.opentelemetry.io/otel` v1.43.0 and `go.opentelemetry.io/otel/trace` v1.43.0 as direct dependencies.
