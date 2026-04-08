package server

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// newTestMeter creates a ManualReader-backed Meter for testing.
func newTestMeter() (metric.Meter, *sdkmetric.ManualReader) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	return mp.Meter("test"), reader
}

// collectMetrics collects all metrics from the reader into a name→Metrics map.
func collectMetrics(t *testing.T, reader *sdkmetric.ManualReader) map[string]metricdata.Metrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}
	result := make(map[string]metricdata.Metrics)
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			result[m.Name] = m
		}
	}
	return result
}

func TestNewWebSocketMetrics(t *testing.T) {
	t.Run("with meter", func(t *testing.T) {
		meter, _ := newTestMeter()
		metrics := NewWebSocketMetrics(meter)

		if metrics == nil {
			t.Fatal("Expected metrics to be non-nil")
		}

		// Verify key metrics are initialized (OTel instruments are never nil)
		if metrics.activeConnections == nil {
			t.Error("activeConnections should be initialized")
		}
		if metrics.totalConnections == nil {
			t.Error("totalConnections should be initialized")
		}
		if metrics.connectionDuration == nil {
			t.Error("connectionDuration should be initialized")
		}
		if metrics.messagesReceived == nil {
			t.Error("messagesReceived should be initialized")
		}
		if metrics.messagesSent == nil {
			t.Error("messagesSent should be initialized")
		}
	})

	t.Run("with nil meter", func(t *testing.T) {
		metrics := NewWebSocketMetrics(nil)
		if metrics != nil {
			t.Error("Expected metrics to be nil when meter is nil")
		}
	})
}

func TestWebSocketMetrics_ConnectionLifecycle(t *testing.T) {
	meter, reader := newTestMeter()
	metrics := NewWebSocketMetrics(meter)
	ctx := context.Background()

	metrics.RecordConnectionStart(ctx)
	metrics.RecordConnectionActive(ctx, 5)
	metrics.RecordConnectionEnd(ctx, 30*time.Second)
	metrics.RecordConnectionError(ctx, "upgrade_failed")

	collected := collectMetrics(t, reader)

	if _, ok := collected["websocket.connections"]; !ok {
		t.Error("Expected websocket.connections metric")
	}
	if _, ok := collected["websocket.active_connections"]; !ok {
		t.Error("Expected websocket.active_connections metric")
	}
	if _, ok := collected["websocket.connection.duration"]; !ok {
		t.Error("Expected websocket.connection.duration metric")
	}
	if _, ok := collected["websocket.connection.errors"]; !ok {
		t.Error("Expected websocket.connection.errors metric")
	}
}

func TestWebSocketMetrics_Messages(t *testing.T) {
	meter, reader := newTestMeter()
	metrics := NewWebSocketMetrics(meter)
	ctx := context.Background()

	metrics.RecordMessageReceived(ctx, 256, "subscribe")
	metrics.RecordMessageSent(ctx, 512, "event")
	metrics.RecordMessageError(ctx, "parse_error", "unknown")

	collected := collectMetrics(t, reader)

	if _, ok := collected["websocket.received.messages"]; !ok {
		t.Error("Expected websocket.received.messages metric")
	}
	if _, ok := collected["websocket.sent.messages"]; !ok {
		t.Error("Expected websocket.sent.messages metric")
	}
	if _, ok := collected["websocket.message.errors"]; !ok {
		t.Error("Expected websocket.message.errors metric")
	}
	if _, ok := collected["websocket.message.size"]; !ok {
		t.Error("Expected websocket.message.size metric")
	}
}

func TestWebSocketMetrics_Requests(t *testing.T) {
	meter, reader := newTestMeter()
	metrics := NewWebSocketMetrics(meter)
	ctx := context.Background()

	// Test successful request
	recordCompletion := metrics.RecordRequest(ctx, "subscribe")
	time.Sleep(10 * time.Millisecond)
	recordCompletion(nil)

	// Test request with error
	recordCompletion2 := metrics.RecordRequest(ctx, "event")
	recordCompletion2(errors.New("topic required"))

	collected := collectMetrics(t, reader)

	if _, ok := collected["websocket.requests"]; !ok {
		t.Error("Expected websocket.requests metric")
	}
	if _, ok := collected["websocket.request.duration"]; !ok {
		t.Error("Expected websocket.request.duration metric")
	}
	if _, ok := collected["websocket.request.errors"]; !ok {
		t.Error("Expected websocket.request.errors metric")
	}
}

func TestWebSocketMetrics_Health(t *testing.T) {
	meter, reader := newTestMeter()
	metrics := NewWebSocketMetrics(meter)
	ctx := context.Background()

	metrics.RecordPingSent(ctx)
	metrics.RecordPongTimeout(ctx)
	metrics.RecordWriteTimeout(ctx)

	collected := collectMetrics(t, reader)

	if _, ok := collected["websocket.pings_sent"]; !ok {
		t.Error("Expected websocket.pings_sent metric")
	}
	if _, ok := collected["websocket.pong_timeouts"]; !ok {
		t.Error("Expected websocket.pong_timeouts metric")
	}
	if _, ok := collected["websocket.write_timeouts"]; !ok {
		t.Error("Expected websocket.write_timeouts metric")
	}
}

func TestWebSocketMetrics_NilSafety(t *testing.T) {
	var metrics *WebSocketMetrics
	ctx := context.Background()

	// These should not panic
	metrics.RecordConnectionStart(ctx)
	metrics.RecordConnectionActive(ctx, 1)
	metrics.RecordConnectionEnd(ctx, time.Second)
	metrics.RecordConnectionError(ctx, "test")
	metrics.RecordMessageReceived(ctx, 100, "test")
	metrics.RecordMessageSent(ctx, 100, "test")
	metrics.RecordMessageError(ctx, "test", "test")
	metrics.RecordPingSent(ctx)
	metrics.RecordPongTimeout(ctx)
	metrics.RecordWriteTimeout(ctx)

	recordCompletion := metrics.RecordRequest(ctx, "test")
	recordCompletion(nil) // Should not panic
}

func TestWebSocketMetrics_ConcurrentAccess(t *testing.T) {
	meter, reader := newTestMeter()
	metrics := NewWebSocketMetrics(meter)
	ctx := context.Background()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				metrics.RecordConnectionStart(ctx)
				metrics.RecordMessageReceived(ctx, 100, "test")
				metrics.RecordMessageSent(ctx, 100, "test")
				recordCompletion := metrics.RecordRequest(ctx, "test")
				recordCompletion(nil)
			}
		}()
	}

	wg.Wait()

	// Verify metrics were collected without panics
	collected := collectMetrics(t, reader)
	if _, ok := collected["websocket.connections"]; !ok {
		t.Error("Expected websocket.connections metric after concurrent access")
	}
}
