package server

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// WebSocketMetrics defines the standard metrics collected by the WebSocket server.
// This struct holds references to all the OTel metric instruments used for monitoring
// WebSocket server performance and behavior.
type WebSocketMetrics struct {
	// Connection metrics
	activeConnections  metric.Float64Gauge    // websocket.active_connections
	totalConnections   metric.Int64Counter    // websocket.connections
	connectionDuration metric.Float64Histogram // websocket.connection.duration
	connectionErrors   metric.Int64Counter    // websocket.connection.errors

	// Message metrics
	messagesReceived metric.Int64Counter    // websocket.received.messages
	messagesSent     metric.Int64Counter    // websocket.sent.messages
	messageErrors    metric.Int64Counter    // websocket.message.errors
	messageSize      metric.Float64Histogram // websocket.message.size

	// Request metrics (by message kind)
	requestsTotal   metric.Int64Counter    // websocket.requests
	requestDuration metric.Float64Histogram // websocket.request.duration
	requestErrors   metric.Int64Counter    // websocket.request.errors

	// Health metrics
	pingsSent     metric.Int64Counter // websocket.pings_sent
	pongTimeouts  metric.Int64Counter // websocket.pong_timeouts
	writeTimeouts metric.Int64Counter // websocket.write_timeouts
}

// newWebSocketMetricsFromProvider creates WebSocketMetrics from a MeterProvider.
// Returns nil if the provider is nil.
func newWebSocketMetricsFromProvider(mp metric.MeterProvider) *WebSocketMetrics {
	if mp == nil {
		return nil
	}
	return NewWebSocketMetrics(mp.Meter("github.com/tsarna/vinculum-vws/server"))
}

// NewWebSocketMetrics creates a new WebSocketMetrics instance using the provided Meter.
// If the meter is nil, returns nil (no metrics will be collected).
func NewWebSocketMetrics(meter metric.Meter) *WebSocketMetrics {
	if meter == nil {
		return nil
	}

	ac, _ := meter.Float64Gauge("websocket.active_connections",
		metric.WithUnit("{connection}"),
		metric.WithDescription("Current number of active WebSocket connections"),
	)
	tc, _ := meter.Int64Counter("websocket.connections",
		metric.WithUnit("{connection}"),
		metric.WithDescription("Total WebSocket connections established"),
	)
	cd, _ := meter.Float64Histogram("websocket.connection.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of WebSocket connections"),
	)
	ce, _ := meter.Int64Counter("websocket.connection.errors",
		metric.WithUnit("{error}"),
		metric.WithDescription("WebSocket connection errors"),
	)

	mr, _ := meter.Int64Counter("websocket.received.messages",
		metric.WithUnit("{message}"),
		metric.WithDescription("Messages received from WebSocket clients"),
	)
	ms, _ := meter.Int64Counter("websocket.sent.messages",
		metric.WithUnit("{message}"),
		metric.WithDescription("Messages sent to WebSocket clients"),
	)
	me, _ := meter.Int64Counter("websocket.message.errors",
		metric.WithUnit("{error}"),
		metric.WithDescription("WebSocket message processing errors"),
	)
	msz, _ := meter.Float64Histogram("websocket.message.size",
		metric.WithUnit("By"),
		metric.WithDescription("Size of WebSocket messages"),
	)

	rt, _ := meter.Int64Counter("websocket.requests",
		metric.WithUnit("{request}"),
		metric.WithDescription("WebSocket requests by kind"),
	)
	rd, _ := meter.Float64Histogram("websocket.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("WebSocket request processing duration"),
	)
	re, _ := meter.Int64Counter("websocket.request.errors",
		metric.WithUnit("{error}"),
		metric.WithDescription("WebSocket request errors"),
	)

	ps, _ := meter.Int64Counter("websocket.pings_sent",
		metric.WithUnit("{ping}"),
		metric.WithDescription("Ping frames sent to WebSocket clients"),
	)
	pt, _ := meter.Int64Counter("websocket.pong_timeouts",
		metric.WithUnit("{timeout}"),
		metric.WithDescription("Pong timeouts (dead connections)"),
	)
	wt, _ := meter.Int64Counter("websocket.write_timeouts",
		metric.WithUnit("{timeout}"),
		metric.WithDescription("Write operation timeouts"),
	)

	return &WebSocketMetrics{
		activeConnections:  ac,
		totalConnections:   tc,
		connectionDuration: cd,
		connectionErrors:   ce,
		messagesReceived:   mr,
		messagesSent:       ms,
		messageErrors:      me,
		messageSize:        msz,
		requestsTotal:      rt,
		requestDuration:    rd,
		requestErrors:      re,
		pingsSent:          ps,
		pongTimeouts:       pt,
		writeTimeouts:      wt,
	}
}

// Connection lifecycle metrics

// RecordConnectionStart records when a new WebSocket connection is established.
func (m *WebSocketMetrics) RecordConnectionStart(ctx context.Context) {
	if m == nil {
		return
	}
	m.totalConnections.Add(ctx, 1)
}

// RecordConnectionActive updates the active connection count.
func (m *WebSocketMetrics) RecordConnectionActive(ctx context.Context, count int) {
	if m == nil {
		return
	}
	m.activeConnections.Record(ctx, float64(count))
}

// RecordConnectionEnd records when a WebSocket connection ends and its duration.
func (m *WebSocketMetrics) RecordConnectionEnd(ctx context.Context, duration time.Duration) {
	if m == nil {
		return
	}
	m.connectionDuration.Record(ctx, duration.Seconds())
}

// RecordConnectionError records connection-level errors (upgrade failures, etc.).
func (m *WebSocketMetrics) RecordConnectionError(ctx context.Context, errorType string) {
	if m == nil {
		return
	}
	m.connectionErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("error.type", errorType)))
}

// Message metrics

// RecordMessageReceived records when a message is received from a client.
func (m *WebSocketMetrics) RecordMessageReceived(ctx context.Context, sizeBytes int, messageKind string) {
	if m == nil {
		return
	}
	m.messagesReceived.Add(ctx, 1, metric.WithAttributes(attribute.String("websocket.message.kind", messageKind)))
	m.messageSize.Record(ctx, float64(sizeBytes), metric.WithAttributes(attribute.String("websocket.message.direction", "received")))
}

// RecordMessageSent records when a message is sent to a client.
func (m *WebSocketMetrics) RecordMessageSent(ctx context.Context, sizeBytes int, messageType string) {
	if m == nil {
		return
	}
	m.messagesSent.Add(ctx, 1, metric.WithAttributes(attribute.String("websocket.message.type", messageType)))
	m.messageSize.Record(ctx, float64(sizeBytes), metric.WithAttributes(attribute.String("websocket.message.direction", "sent")))
}

// RecordMessageError records message processing errors.
func (m *WebSocketMetrics) RecordMessageError(ctx context.Context, errorType string, messageKind string) {
	if m == nil {
		return
	}
	m.messageErrors.Add(ctx, 1, metric.WithAttributes(
		attribute.String("error.type", errorType),
		attribute.String("websocket.message.kind", messageKind),
	))
}

// Request metrics

// RecordRequest records the start of a request and returns a function to record completion.
// Usage:
//
//	recordCompletion := metrics.RecordRequest(ctx, "subscribe")
//	defer recordCompletion(err)
func (m *WebSocketMetrics) RecordRequest(ctx context.Context, requestKind string) func(error) {
	if m == nil {
		return func(error) {} // No-op function
	}

	startTime := time.Now()
	m.requestsTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("websocket.request.kind", requestKind)))

	return func(err error) {
		duration := time.Since(startTime)
		m.requestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attribute.String("websocket.request.kind", requestKind)))

		if err != nil {
			m.requestErrors.Add(ctx, 1, metric.WithAttributes(
				attribute.String("websocket.request.kind", requestKind),
				attribute.String("error.type", err.Error()),
			))
		}
	}
}

// Health metrics

// RecordPingSent records when a ping frame is sent to a client.
func (m *WebSocketMetrics) RecordPingSent(ctx context.Context) {
	if m == nil {
		return
	}
	m.pingsSent.Add(ctx, 1)
}

// RecordPongTimeout records when a client fails to respond to a ping (dead connection).
func (m *WebSocketMetrics) RecordPongTimeout(ctx context.Context) {
	if m == nil {
		return
	}
	m.pongTimeouts.Add(ctx, 1)
}

// RecordWriteTimeout records when a write operation times out.
func (m *WebSocketMetrics) RecordWriteTimeout(ctx context.Context) {
	if m == nil {
		return
	}
	m.writeTimeouts.Add(ctx, 1)
}
