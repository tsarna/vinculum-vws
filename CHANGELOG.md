# Changelog

## [Unreleased]

## [0.14.1] - 2026-08-25

Same change as 0.14.0, released on top of the dependency updates 0.14.0 was
tagged without. **Use this instead of 0.14.0**, which was published from a
commit that predates the `vinculum-bus` 0.16.0 bump. The bad tag is left in
place rather than moved: a tag the module proxy has already served cannot be
changed without poisoning that version.

### Added

- **`Client.IsConnected()`** reports whether the client currently holds a live
  WebSocket connection. It exists for health reporting: a host answering a
  readiness probe needs to say "this process cannot do its job right now" while
  the connection is down.

  It is a snapshot, not a guarantee — the connection may drop between the call
  and the next `Publish` — so it is useful for a probe and useless as a
  precondition. It is deliberately stricter than the check the operational
  methods make, requiring both the started flag and the connection itself, so
  the brief window inside cleanup where the socket is already closed reads as
  disconnected.

## [0.13.0] - 2026-05-25

Change license to Apache-2.0

## [0.12.0] - 2026-04-23

### Changed

Adapt to API change for transforms in vinculum-bus 0.14.0

## [0.11.2] - 2026-04-23

### Changed

- **Topic matching routes through `vinculum-bus/topicmatch`** — the event-authorization allow-pattern check now honors MQTT 5.0 §4.7.2: filters starting with `+` or `#` no longer match reserved `$`-prefixed topics. Exact and `$`-prefixed patterns are unaffected. Requires vinculum-bus v0.12.0.

## [0.11.1] - 2026-04-18

### Added

- **`vinculum.server.name` metric attribute** — all server metrics now carry a `vinculum.server.name` attribute identifying the vinculum server block. The listener config accepts `WithServerName(name)` to set it, and `NewWebSocketMetrics` now takes a `serverName` parameter.

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
